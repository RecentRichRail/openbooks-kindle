package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evan-buss/openbooks/core"
	"github.com/evan-buss/openbooks/util"
)

// RequestHandler defines a generic handle() method that is called when a specific request type is made
type RequestHandler interface {
	handle(c *Client)
}

// messageRouter is used to parse the incoming request and respond appropriately
func (server *server) routeMessage(message Request, c *Client) {
	var obj interface{}

	switch message.MessageType {
	case SEARCH:
		obj = new(SearchRequest)
	case DOWNLOAD:
		obj = new(DownloadRequest)
	case SEND_TO_KINDLE:
		obj = new(SendToKindleRequest)
	}

	err := json.Unmarshal(message.Payload, &obj)
	if err != nil {
		server.log.Printf("Invalid request payload. %s.\n", err.Error())
		c.send <- StatusResponse{
			MessageType:      STATUS,
			NotificationType: DANGER,
			Title:            "Unknown request payload.",
		}
	}

	switch message.MessageType {
	case CONNECT:
		c.startIrcConnection(server)
	case SEARCH:
		c.sendSearchRequest(obj.(*SearchRequest), server)
	case DOWNLOAD:
		c.sendDownloadRequest(obj.(*DownloadRequest))
	case SEND_TO_KINDLE:
		c.sendToKindle(obj.(*SendToKindleRequest), server)
	default:
		server.log.Println("Unknown request type received.")
	}
}

// handle ConnectionRequests and either connect to the server or do nothing
func (c *Client) startIrcConnection(server *server) {
	err := core.Join(c.irc, server.config.Server, server.config.EnableTLS)
	if err != nil {
		c.log.Println(err)
		c.send <- newErrorResponse("Unable to connect to IRC server.")
		return
	}

	handler := server.NewIrcEventHandler(c)

	if server.config.Log {
		logger, _, err := util.CreateLogFile(c.irc.Username, server.config.DownloadDir)
		if err != nil {
			server.log.Println(err)
		}
		handler[core.Message] = func(text string) { logger.Println(text) }
	}

	go core.StartReader(c.ctx, c.irc, handler)

	c.send <- ConnectionResponse{
		StatusResponse: StatusResponse{
			MessageType:      CONNECT,
			NotificationType: SUCCESS,
			Title:            "Welcome, connection established.",
			Detail:           fmt.Sprintf("IRC username %s", c.irc.Username),
		},
		Name: c.irc.Username,
	}
}

// handle SearchRequests and send the query to the book server
func (c *Client) sendSearchRequest(s *SearchRequest, server *server) {
	server.lastSearchMutex.Lock()
	defer server.lastSearchMutex.Unlock()

	nextAvailableSearch := server.lastSearch.Add(server.config.SearchTimeout)

	if time.Now().Before(nextAvailableSearch) {
		remainingSeconds := time.Until(nextAvailableSearch).Seconds()
		c.send <- newRateLimitResponse(remainingSeconds)

		return
	}

	core.SearchBook(c.irc, server.config.SearchBot, s.Query)
	server.lastSearch = time.Now()

	c.send <- newStatusResponse(NOTIFY, "Search request sent.")
}

// handle DownloadRequests by sending the request to the book server
func (c *Client) sendDownloadRequest(d *DownloadRequest) {
	core.DownloadBook(c.irc, d.Book)
	c.send <- newStatusResponse(NOTIFY, "Download request received.")
}

// handle SendToKindleRequests by downloading the book and emailing it
func (c *Client) sendToKindle(req *SendToKindleRequest, server *server) {
	// Debug: Log the book details being requested
	server.log.Printf("SERVER: Send to Kindle request for book: %+v", req.Book)
	server.log.Printf("SERVER: Email address: %s", req.Email)
	
	if !server.config.SMTPEnabled {
		c.send <- newStatusResponse(WARNING, "Email functionality is not configured. Please check SMTP settings.")
		return
	}
	
	// Send the download request to IRC
	server.log.Printf("SERVER: Sending IRC download request...")
	core.DownloadBook(c.irc, req.Book)
	
	c.send <- newStatusResponse(NOTIFY, "Download request sent. Waiting for book to download...")
	
	go func() {
		// Wait for the download to complete and find the file
		downloadedFilePath := ""
		maxWaitTime := 5 * time.Minute
		checkInterval := 2 * time.Second
		startTime := time.Now()
		
		server.log.Printf("SERVER: Starting download monitoring, will check every %v for up to %v", checkInterval, maxWaitTime)
		
		for time.Since(startTime) < maxWaitTime {
			elapsed := time.Since(startTime)
			checkNum := int(elapsed/checkInterval) + 1
			server.log.Printf("SERVER: Download check #%d (elapsed: %v)", checkNum, elapsed.Round(time.Second))
			
			// Check for newly downloaded files in the books directory
			booksDir := filepath.Join(server.config.DownloadDir, "books")
			files, err := os.ReadDir(booksDir)
			if err != nil {
				server.log.Printf("SERVER: Error reading books directory: %v", err)
				time.Sleep(checkInterval)
				continue
			}
			
			server.log.Printf("SERVER: Found %d files in books directory", len(files))
			
			// Look for files created within the last 2 minutes (since our request started)
			cutoffTime := startTime.Add(-1 * time.Minute) // Allow some buffer
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				
				filePath := filepath.Join(booksDir, file.Name())
				fileInfo, err := file.Info()
				if err != nil {
					continue
				}
				
				// Check if this file was modified recently and is not a temporary file
				if fileInfo.ModTime().After(cutoffTime) && !strings.HasSuffix(file.Name(), ".temp") {
					server.log.Printf("SERVER: Found recent file: %s (modified: %v)", file.Name(), fileInfo.ModTime())
					
					// Additional check: make sure the file is complete (not being written to)
					// Wait a moment and check if the size is stable
					initialSize := fileInfo.Size()
					time.Sleep(1 * time.Second)
					
					if updatedInfo, err := os.Stat(filePath); err == nil {
						if updatedInfo.Size() == initialSize && updatedInfo.Size() > 1000 { // File is stable and has reasonable size
							downloadedFilePath = filePath
							break
						}
					}
				}
			}
			
			if downloadedFilePath != "" {
				break
			}
			
			time.Sleep(checkInterval)
		}
		
		if downloadedFilePath == "" {
			server.log.Printf("SERVER: No downloaded file found after %v", maxWaitTime)
			c.send <- newStatusResponse(DANGER, "Download timed out or failed. The book may not be available.")
			return
		}
		
		server.log.Printf("SERVER: Download completed, file found: %s", downloadedFilePath)
		c.send <- newStatusResponse(NOTIFY, "Book downloaded! Sending to "+req.Email+"...")
		
		// Extract title and author from the book request or filename
		title := "Unknown Title"
		author := "Unknown Author"
		if req.Book != "" {
			// Try to parse title/author from book request string
			// Book strings are typically in format "Title by Author" 
			parts := strings.Split(req.Book, " by ")
			if len(parts) >= 2 {
				title = strings.TrimSpace(parts[0])
				author = strings.TrimSpace(parts[1])
			} else {
				title = req.Book
			}
		}
		
		// Send the book via email
		err := server.sendBookViaEmail(req.Email, title, author, downloadedFilePath)
		if err != nil {
			server.log.Printf("SERVER: Email sending failed: %v", err)
			c.send <- newStatusResponse(DANGER, fmt.Sprintf("Failed to send email: %v", err))
		} else {
			server.log.Printf("SERVER: Email sent successfully to %s", req.Email)
			c.send <- newStatusResponse(SUCCESS, "Book sent to your email successfully!")
			
			// Clean up the downloaded file
			if err := os.Remove(downloadedFilePath); err != nil {
				server.log.Printf("SERVER: Failed to clean up file %s: %v", downloadedFilePath, err)
			}
		}
	}()
}
