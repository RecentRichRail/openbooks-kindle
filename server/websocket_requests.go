package server

import (
	"encoding/json"
	"fmt"
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
	go func() {
		if !server.config.SMTPEnabled {
			c.send <- newStatusResponse(WARNING, "Email functionality is not configured. Please check SMTP settings.")
			return
		}
		
		// Download the book
		c.send <- newStatusResponse(NOTIFY, "Downloading book for Kindle...")
		
		// This will download the book to the server's download directory
		core.DownloadBook(c.irc, req.Book)
		
		// Send book via email
		c.send <- newStatusResponse(NOTIFY, "Sending book to "+req.Email+"...")
		
		// TODO: Get the actual downloaded file path and send it via email
		// For now, just send a success message
		err := server.sendBookViaEmail(req.Email, req.Title, req.Author, req.Book)
		if err != nil {
			server.log.Printf("Failed to send book via email: %s", err)
			c.send <- newStatusResponse(DANGER, "Failed to send book to Kindle: "+err.Error())
		} else {
			c.send <- newStatusResponse(SUCCESS, "Book sent to "+req.Email+" successfully!")
		}
	}()
}
