package middleware

import (
	"context"
	"net/http"
	"strings"

	"golang-firebase-backend/config"
	"golang-firebase-backend/utils"
)

func FirebaseAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondError(w, http.StatusUnauthorized, "Missing or invalid token")
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == "" {
			utils.RespondError(w, http.StatusUnauthorized, "Token is required")
			return
		}

		// Initialize Firebase Auth client
		ctx := context.Background()
		client, err := config.FirebaseApp.Auth(ctx)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to initialize Firebase Auth")
			return
		}

		// Verify the ID token
		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			utils.RespondError(w, http.StatusUnauthorized, "Invalid ID Token")
			return
		}

		// Add UID to context and pass it to the next handler
		uid := token.UID
		ctx = context.WithValue(r.Context(), "uid", uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
