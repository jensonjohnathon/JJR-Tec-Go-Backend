package server

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware checks the Authorization header for a valid JWT token.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")

        if tokenString == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        // Remove "Bearer " prefix if it exists
        tokenString = strings.TrimPrefix(tokenString, "Bearer ")

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtKey), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Token is valid; proceed with the request
        next.ServeHTTP(w, r)
    })
}
