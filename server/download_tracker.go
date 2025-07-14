package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DownloadStatus represents the status of a download
type DownloadStatus string

const (
	DownloadPending    DownloadStatus = "pending"
	DownloadStarted    DownloadStatus = "started"
	DownloadProgress   DownloadStatus = "progress"
	DownloadCompleted  DownloadStatus = "completed"
	DownloadFailed     DownloadStatus = "failed"
)

// DownloadInfo tracks information about a download
type DownloadInfo struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Author      string         `json:"author"`
	BookCommand string         `json:"bookCommand"`
	Status      DownloadStatus `json:"status"`
	Progress    int            `json:"progress"`     // 0-100
	FilePath    string         `json:"filePath"`    // Path to downloaded file
	FileName    string         `json:"fileName"`    // Name of downloaded file
	StartTime   time.Time      `json:"startTime"`
	EndTime     *time.Time     `json:"endTime,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// DownloadTracker manages download progress and file tracking
type DownloadTracker struct {
	downloads map[string]*DownloadInfo
	mutex     sync.RWMutex
	baseDir   string
}

// NewDownloadTracker creates a new download tracker
func NewDownloadTracker(baseDir string) *DownloadTracker {
	booksDir := filepath.Join(baseDir, "books")
	
	// Ensure the books directory exists
	err := os.MkdirAll(booksDir, 0755)
	if err != nil {
		log.Printf("Error creating books directory: %v", err)
	}
	
	log.Printf("TRACKER INIT: Base directory: %s", baseDir)
	log.Printf("TRACKER INIT: Books directory: %s", booksDir)
	log.Printf("TRACKER INIT: Download tracker initialized")
	
	return &DownloadTracker{
		downloads: make(map[string]*DownloadInfo),
		baseDir:   baseDir,
	}
}

// StartDownload creates a new download tracking entry
func (dt *DownloadTracker) StartDownload(id, title, author, bookCommand string) *DownloadInfo {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	log.Printf("TRACKER: Starting download tracking for ID: %s, Book: %s by %s", id, title, author)

	download := &DownloadInfo{
		ID:          id,
		Title:       title,
		Author:      author,
		BookCommand: bookCommand,
		Status:      DownloadStarted,
		Progress:    0,
		StartTime:   time.Now(),
	}

	dt.downloads[id] = download
	log.Printf("TRACKER: Download %s registered and tracking started", id)
	return download
}

// UpdateProgress updates the progress of a download
func (dt *DownloadTracker) UpdateProgress(id string, progress int) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	if download, exists := dt.downloads[id]; exists {
		download.Progress = progress
		if progress >= 100 {
			download.Status = DownloadCompleted
			now := time.Now()
			download.EndTime = &now
			
			// Try to find the downloaded file
			dt.findDownloadedFile(download)
		} else {
			download.Status = DownloadProgress
		}
	}
}

// MarkFailed marks a download as failed
func (dt *DownloadTracker) MarkFailed(id string, errorMsg string) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	if download, exists := dt.downloads[id]; exists {
		download.Status = DownloadFailed
		download.Error = errorMsg
		now := time.Now()
		download.EndTime = &now
	}
}

// GetDownload retrieves download information
func (dt *DownloadTracker) GetDownload(id string) (*DownloadInfo, bool) {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()

	download, exists := dt.downloads[id]
	return download, exists
}

// GetAllDownloads returns all downloads
func (dt *DownloadTracker) GetAllDownloads() map[string]*DownloadInfo {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*DownloadInfo)
	for k, v := range dt.downloads {
		downloadCopy := *v
		result[k] = &downloadCopy
	}
	return result
}

// TriggerFileDetection manually triggers file detection for a download
func (dt *DownloadTracker) TriggerFileDetection(id string) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	if download, exists := dt.downloads[id]; exists {
		dt.findDownloadedFile(download)
	}
}

// findDownloadedFile attempts to locate the downloaded file
func (dt *DownloadTracker) findDownloadedFile(download *DownloadInfo) {
	booksDir := filepath.Join(dt.baseDir, "books")
	
	// List all files in the directory for debugging
	files, err := os.ReadDir(booksDir)
	if err != nil {
		log.Printf("TRACKER: Error reading books directory %s: %v", booksDir, err)
		return
	}
	
	log.Printf("TRACKER: Looking for download %s in directory: %s", download.ID, booksDir)
	log.Printf("TRACKER: Found %d files in books directory:", len(files))
	for _, file := range files {
		if !file.IsDir() {
			info, _ := file.Info()
			log.Printf("  - %s (modified: %v)", file.Name(), info.ModTime())
		}
	}
	
	// Generate potential filenames
	potentialNames := dt.generatePotentialFilenames(download.Title, download.Author)
	
	// First try to find by exact filename match
	for _, name := range potentialNames {
		fullPath := filepath.Join(booksDir, name)
		if _, err := os.Stat(fullPath); err == nil {
			download.FilePath = fullPath
			download.FileName = name
			log.Printf("Found file by name match: %s", fullPath)
			return
		}
	}
	
	// If not found by name, look for the most recent file
	// files already declared above, just check for errors
	if err != nil {
		return
	}
	
	var newestFile os.DirEntry
	var newestTime int64
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		// Skip test files and temporary files
		name := file.Name()
		if contains(name, "test-book") || hasSuffix(name, ".temp") {
			continue
		}
		
		info, err := file.Info()
		if err != nil {
			continue
		}
		
		// Only consider files modified after download started
		if info.ModTime().After(download.StartTime) && info.ModTime().Unix() > newestTime {
			newestTime = info.ModTime().Unix()
			newestFile = file
		}
	}
	
	if newestFile != nil {
		download.FilePath = filepath.Join(booksDir, newestFile.Name())
		download.FileName = newestFile.Name()
		log.Printf("TRACKER: Found downloaded file: %s for download ID: %s", download.FilePath, download.ID)
	} else {
		log.Printf("TRACKER: No downloaded file found yet for download ID: %s", download.ID)
	}
}

// generatePotentialFilenames creates potential filenames for a book
func (dt *DownloadTracker) generatePotentialFilenames(title, author string) []string {
	cleanTitle := sanitizeFilename(title)
	cleanAuthor := sanitizeFilename(author)
	
	extensions := []string{".epub", ".pdf", ".mobi", ".txt", ".azw", ".azw3", ".fb2", ".djvu"}
	formats := []string{
		fmt.Sprintf("%s - %s", cleanTitle, cleanAuthor),
		fmt.Sprintf("%s_%s", cleanTitle, cleanAuthor),
		fmt.Sprintf("%s by %s", cleanTitle, cleanAuthor),
		cleanTitle,
		fmt.Sprintf("%s %s", cleanTitle, cleanAuthor),
	}
	
	var filenames []string
	for _, format := range formats {
		for _, ext := range extensions {
			filenames = append(filenames, format+ext)
		}
		// Also try without extension
		filenames = append(filenames, format)
	}
	
	return filenames
}

// CleanupOldDownloads removes old download entries (older than 24 hours)
func (dt *DownloadTracker) CleanupOldDownloads() {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for id, download := range dt.downloads {
		if download.StartTime.Before(cutoff) {
			delete(dt.downloads, id)
		}
	}
}

// sanitizeFilename removes invalid characters from filenames
func sanitizeFilename(filename string) string {
	// Remove or replace invalid characters for filenames
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	result := filename
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Trim spaces and dots from beginning and end
	result = strings.Trim(result, " .")
	// Limit length to prevent filesystem issues
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}

// Helper functions
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
