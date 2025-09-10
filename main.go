package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"twitter-down/handlers"
	"twitter-down/middleware"
	"twitter-down/utils"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.JSONResponse(w, true, "API is running", []string{
			"/ - root endpoint",
			"/twitter?url={tweet_url} - Download images from a tweet",
		})
	})

	// API Endpoints with CORS middleware
	mux.Handle("/generic", middleware.CORS(handlers.GenericDownloadHandler()))
	mux.Handle("/twitter", middleware.CORS(handlers.TwitterDownloadHandler()))
	mux.Handle("/pinterest", middleware.CORS(handlers.PinterestDownloadHandler()))
	mux.Handle("/instagram", middleware.CORS(handlers.InstagramDownloadHandler()))

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	fmt.Printf("Server running at http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
