package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/cors"
)

type server struct {
	// Shared app configuration
	config *Config

	// Shared data
	repository *Repository

	// SMTP service for sending emails
	smtpService *SMTPService

	// Registered clients.
	clients map[uuid.UUID]*Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	log *log.Logger

	// Mutex to guard the lastSearch timestamp
	lastSearchMutex sync.Mutex

	// The time the last search was performed. Used to rate limit searches.
	lastSearch time.Time
}

// Config contains settings for server
type Config struct {
	Log                     bool
	Port                    string
	UserName                string
	Persist                 bool
	DownloadDir             string
	Basepath                string
	Server                  string
	EnableTLS               bool
	SearchTimeout           time.Duration
	SearchBot               string
	DisableBrowserDownloads bool
	UserAgent               string
	// SMTP Configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPEnabled  bool
}

func New(config Config) *server {
	return &server{
		repository:  NewRepository(),
		config:      &config,
		smtpService: NewSMTPService(&config),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[uuid.UUID]*Client),
		log:         log.New(os.Stdout, "SERVER: ", log.LstdFlags|log.Lmsgprefix),
	}
}

// Start instantiates the web server and opens the browser
func Start(config Config) {
	createBooksDirectory(config)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)

	corsConfig := cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"http://127.0.0.1:5173"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "DELETE"},
	}
	router.Use(cors.New(corsConfig).Handler)

	server := New(config)
	routes := server.registerRoutes()

	ctx, cancel := context.WithCancel(context.Background())
	go server.startClientHub(ctx)
	server.registerGracefulShutdown(cancel)
	router.Mount(config.Basepath, routes)

	server.log.Printf("Base Path: %s\n", config.Basepath)
	server.log.Printf("OpenBooks is listening on port %v", config.Port)
	server.log.Printf("Download Directory: %s\n", config.DownloadDir)
	server.log.Printf("Open http://localhost:%v%s in your browser.", config.Port, config.Basepath)
	server.log.Fatal(http.ListenAndServe(":"+config.Port, router))
}

// The client hub is to be run in a goroutine and handles management of
// websocket client registrations.
func (server *server) startClientHub(ctx context.Context) {
	for {
		select {
		case client := <-server.register:
			server.clients[client.uuid] = client
		case client := <-server.unregister:
			if _, ok := server.clients[client.uuid]; ok {
				_, cancel := context.WithCancel(client.ctx)
				close(client.send)
				cancel()
				delete(server.clients, client.uuid)
			}
		case <-ctx.Done():
			for _, client := range server.clients {
				_, cancel := context.WithCancel(client.ctx)
				close(client.send)
				cancel()
				delete(server.clients, client.uuid)
			}
			return
		}
	}
}

func (server *server) registerGracefulShutdown(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		server.log.Println("Graceful shutdown.")
		// Close the shutdown channel. Triggering all reader/writer WS handlers to close.
		cancel()
		time.Sleep(time.Second)
		os.Exit(0)
	}()
}

func createBooksDirectory(config Config) {
	err := os.MkdirAll(filepath.Join(config.DownloadDir, "books"), os.FileMode(0755))
	if err != nil {
		panic(err)
	}
}

// sendBookViaEmail sends a book file to the specified email address
func (s *server) sendBookViaEmail(email, title, author, bookIdentifier string) error {
	if !s.config.SMTPEnabled {
		return fmt.Errorf("SMTP is not enabled")
	}

	// For now, we'll create a mock file to test email functionality
	// In the future, this should find the actual downloaded book file
	bookPath := filepath.Join(s.config.DownloadDir, "books", "test-book.txt")

	// Create a test file if it doesn't exist
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		testContent := fmt.Sprintf("Test book: %s by %s\n\nThis is a test file for SMTP functionality.", title, author)
		err = os.WriteFile(bookPath, []byte(testContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to create test book file: %w", err)
		}
	}

	// Open the book file
	file, err := os.Open(bookPath)
	if err != nil {
		return fmt.Errorf("failed to open book file: %w", err)
	}
	defer file.Close()

	// Send via SMTP
	filename := fmt.Sprintf("%s - %s.txt", title, author)
	return s.smtpService.SendBookToKindle(email, title, author, file, filename)
}
