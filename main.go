package main

import (
	"log"
	"net/http"
	"os"

	"golang-firebase-backend/config"
	"golang-firebase-backend/controllers"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Firebase
	config.InitializeFirebaseApp()

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8100") // Replace with your frontend's origin
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token, Authorization")

			// Handle preflight (OPTIONS) requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Define routes
	mux.HandleFunc("/register", controllers.RegisterWithEmail)
	mux.HandleFunc("/login/email", controllers.LoginWithEmail)
	mux.HandleFunc("/login/google", controllers.LoginWithGoogle)

	// Wrap ServeMux with CORS middleware
	handler := corsMiddleware(mux)

	// Get server port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Start server
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
