package main

import (
	"golang-firebase-backend/config"
	"golang-firebase-backend/controllers"
	"golang-firebase-backend/handlers"
	"golang-firebase-backend/middleware"
	"log"
	"net/http"
	"os" // Import the gorilla mux package

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
	mux.HandleFunc("/login/google", controllers.LoginWithGoogle)
	mux.HandleFunc("/logout", controllers.Logout)

	//user update
	mux.Handle("/user/update", middleware.FirebaseAuthMiddleware(http.HandlerFunc(controllers.UpdateUser)))

	//skill route
	mux.HandleFunc("/skills/fetch", controllers.FetchSkills)
	mux.HandleFunc("/skills/view", controllers.ShowSkill)
	//skillroute for admin
	mux.HandleFunc("/skills/admincreate", controllers.CreateSkill)
	mux.HandleFunc("/skills/adminupdate", controllers.UpdateSkill)
	mux.HandleFunc("/skills/admindelete", controllers.DeleteSkill)
	//userskill routes
	mux.Handle("/skills/add", middleware.FirebaseAuthMiddleware(http.HandlerFunc(controllers.AddUserSkill)))
	mux.Handle("/user/portfolios/view", middleware.FirebaseAuthMiddleware(http.HandlerFunc(controllers.ViewUserPortfolios)))
	mux.Handle("/user/portfolios/view-specific", middleware.FirebaseAuthMiddleware(http.HandlerFunc(controllers.ViewSpecificUserPortfolios)))
	mux.Handle("/user/portfolios/create", middleware.FirebaseAuthMiddleware(http.HandlerFunc(controllers.CreatePortfolio)))
	mux.Handle("/user/portfolios/update", middleware.FirebaseAuthMiddleware(http.HandlerFunc(controllers.UpdatePortfolio)))
	mux.Handle("/user/portfolios/delete", middleware.FirebaseAuthMiddleware(http.HandlerFunc(controllers.DeletePortfolio)))

	mux.Handle("/user/request-seller", middleware.FirebaseAuthMiddleware(http.HandlerFunc(handlers.HandleRequestSeller)))
	mux.Handle("/admin/verify-seller", middleware.FirebaseAuthMiddleware(http.HandlerFunc(handlers.HandleAdminVerifySeller)))
	mux.Handle("/user/change-role", middleware.FirebaseAuthMiddleware(http.HandlerFunc(handlers.HandleChangeRole)))

	//tambahin role admin, seller, buyer sebagai middleware

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
