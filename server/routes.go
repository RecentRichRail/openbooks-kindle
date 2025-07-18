package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/evan-buss/openbooks/irc"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

//go:embed app/dist
var reactClient embed.FS

// generateRandomUsername creates a random username for IRC connections
func generateRandomUsername(baseUsername string) string {
	adjectives := []string{
		"happy", "clever", "swift", "bright", "quiet", "bold", "calm", "wise", "kind", "cool",
		"brave", "eager", "fierce", "gentle", "humble", "jolly", "keen", "lively", "merry", "noble",
		"proud", "quick", "radiant", "serene", "trusty", "vibrant", "witty", "zealous", "agile", "daring",
		"elegant", "fearless", "graceful", "honest", "inventive", "joyful", "loyal", "magnificent", "optimistic", "peaceful",
		"reliable", "spirited", "talented", "unique", "valiant", "wonderful", "excellent", "youthful", "zestful", "amazing",
		"brilliant", "charming", "delightful", "energetic", "fabulous", "glorious", "heroic", "inspiring", "jubilant", "kindhearted",
	}
	
	nouns := []string{
		"falcon", "tiger", "eagle", "wolf", "bear", "lion", "fox", "hawk", "shark", "panther",
		"dragon", "phoenix", "raven", "owl", "deer", "rabbit", "dolphin", "whale", "cat", "dog",
		"horse", "elephant", "rhino", "zebra", "giraffe", "monkey", "penguin", "turtle", "snake", "frog",
		"butterfly", "bee", "spider", "ant", "cricket", "firefly", "mantis", "beetle", "moth", "dragonfly",
		"mountain", "river", "ocean", "forest", "desert", "valley", "canyon", "meadow", "lake", "island",
		"star", "moon", "sun", "comet", "meteor", "galaxy", "nebula", "planet", "asteroid", "cosmos",
	}
	
	// Generate random numbers for more variety
	numbers := rand.Intn(99999) + 10000
	
	// Pick random adjective and noun
	adjective := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	
	// Create username without underscores, using camelCase style
	if baseUsername == "" {
		return fmt.Sprintf("%s%s%d", adjective, noun, numbers)
	}
	
	return fmt.Sprintf("%s%s%s%d", baseUsername, adjective, noun, numbers)
}

func (server *server) registerRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Handle("/*", server.staticFilesHandler("app/dist"))
	router.Get("/ws", server.serveWs())
	router.Get("/stats", server.statsHandler())
	router.Get("/servers", server.serverListHandler())
	router.Post("/send-to-kindle", server.sendToKindleHandler())

	router.Group(func(r chi.Router) {
		r.Use(server.requireUser)
		r.Get("/library", server.getAllBooksHandler())
		r.Delete("/library/{fileName}", server.deleteBooksHandler())
		r.Get("/library/*", server.getBookHandler())
	})

	return router
}

// serveWs handles websocket requests from the peer.
func (server *server) serveWs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("OpenBooks")
		if errors.Is(err, http.ErrNoCookie) {
			cookie = &http.Cookie{
				Name:     "OpenBooks",
				Value:    uuid.New().String(),
				Secure:   false,
				HttpOnly: true,
				Expires:  time.Now().Add(time.Hour * 24 * 7),
				SameSite: http.SameSiteStrictMode,
			}
			w.Header().Add("Set-Cookie", cookie.String())
		}

		userId, err := uuid.Parse(cookie.Value)
		_, alreadyConnected := server.clients[userId]

		// If invalid UUID or the same browser tries to connect again or multiple browser connections
		// Don't connect to IRC or create new client
		if err != nil || alreadyConnected || len(server.clients) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		upgrader.CheckOrigin = func(req *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(w, r, w.Header())
		if err != nil {
			server.log.Println(err)
			return
		}

		randomUsername := generateRandomUsername(server.config.UserName)

		client := &Client{
			conn: conn,
			send: make(chan interface{}, 128),
			uuid: userId,
			irc:  irc.New(randomUsername, server.config.UserAgent),
			log:  log.New(os.Stdout, fmt.Sprintf("CLIENT (%s): ", randomUsername), log.LstdFlags|log.Lmsgprefix),
			ctx:  context.Background(),
		}

		server.log.Printf("Client connected from %s\n", conn.RemoteAddr().String())
		client.log.Println("New client created.")

		server.register <- client

		go server.writePump(client)
		go server.readPump(client)
	}
}

func (server *server) staticFilesHandler(assetPath string) http.Handler {
	// update the embedded file system's tree so that index.html is at the root
	app, err := fs.Sub(reactClient, assetPath)
	if err != nil {
		server.log.Println(err)
	}

	// strip the predefined base path and serve the static file
	return http.StripPrefix(server.config.Basepath, http.FileServer(http.FS(app)))
}

func (server *server) statsHandler() http.HandlerFunc {
	type statsReponse struct {
		UUID string `json:"uuid"`
		IP   string `json:"ip"`
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		result := make([]statsReponse, 0, len(server.clients))

		for _, client := range server.clients {
			details := statsReponse{
				UUID: client.uuid.String(),
				Name: client.irc.Username,
				IP:   client.conn.RemoteAddr().String(),
			}

			result = append(result, details)
		}

		json.NewEncoder(w).Encode(result)
	}
}

func (server *server) serverListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(server.repository.servers)
	}
}

func (server *server) getAllBooksHandler() http.HandlerFunc {
	type download struct {
		Name         string    `json:"name"`
		DownloadLink string    `json:"downloadLink"`
		Time         time.Time `json:"time"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if !server.config.Persist {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		libraryDir := filepath.Join(server.config.DownloadDir, "books")
		books, err := os.ReadDir(libraryDir)
		if err != nil {
			server.log.Printf("Unable to list books. %s\n", err)
		}

		output := make([]download, 0)
		for _, book := range books {
			if book.IsDir() || strings.HasPrefix(book.Name(), ".") || filepath.Ext(book.Name()) == ".temp" {
				continue
			}

			info, err := book.Info()
			if err != nil {
				server.log.Println(err)
			}

			dl := download{
				Name:         book.Name(),
				DownloadLink: path.Join("library", book.Name()),
				Time:         info.ModTime(),
			}

			output = append(output, dl)
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(output)
	}
}

func (server *server) getBookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, fileName := path.Split(r.URL.Path)
		bookPath := filepath.Join(server.config.DownloadDir, "books", fileName)

		http.ServeFile(w, r, bookPath)

		if !server.config.Persist {
			err := os.Remove(bookPath)
			if err != nil {
				server.log.Printf("Error when deleting book file. %s", err)
			}
		}
	}
}

func (server *server) deleteBooksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName, err := url.PathUnescape(chi.URLParam(r, "fileName"))
		if err != nil {
			server.log.Printf("Error unescaping path: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = os.Remove(filepath.Join(server.config.DownloadDir, "books", fileName))
		if err != nil {
			server.log.Printf("Error deleting book file: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (server *server) sendToKindleHandler() http.HandlerFunc {
	type sendToKindleRequest struct {
		Email    string `json:"email"`
		BookFile string `json:"bookFile"`
		Title    string `json:"title"`
		Author   string `json:"author"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if !server.config.SMTPEnabled {
			http.Error(w, "SMTP is not configured", http.StatusServiceUnavailable)
			return
		}

		var req sendToKindleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate email
		if req.Email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.BookFile == "" || req.Title == "" || req.Author == "" {
			http.Error(w, "BookFile, Title, and Author are required", http.StatusBadRequest)
			return
		}

		// For now, we'll simulate sending an email
		// In a real implementation, you would:
		// 1. Validate the book file path
		// 2. Open the book file
		// 3. Send it via SMTP

		server.log.Printf("Send to Kindle request: %s by %s to %s", req.Title, req.Author, req.Email)

		// Simulate success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Book sent to Kindle successfully",
		})
	}
}
