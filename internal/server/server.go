package server

import (
	"log"
	"net/http"

	"github.com/theHimanshuShekhar/bchat-server/internal/config"
	"github.com/theHimanshuShekhar/bchat-server/pkg/logger"
)

// Start initializes and starts the HTTP server
func Start() {
	
	// Load configuration
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize home route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {	
		w.Write([]byte("BChat Server"))
	})

	// Initialize auth route
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {	
		w.Write([]byte("Authentication Route"))
	})

    // Initialize websocket server
    http.HandleFunc("/ws", HandleConnections)

	port := config.ServerPort

    // Start the HTTP server on the specified port
    logger.Info("Starting server on port " + port)
    err = http.ListenAndServe(":" + port, nil)
    if err != nil {
        logger.Error("Failed to start server: ", err)
        log.Fatalf("ListenAndServe: %v", err)
    }
}
