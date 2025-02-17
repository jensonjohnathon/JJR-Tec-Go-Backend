package server

import (
	"net/http"
	"strings"
	"time"

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

        // Parse the JWT token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtKey), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Extract claims from the token
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        // Check if the token has expired
        exp, expOk := claims["exp"].(float64)
        if !expOk {
            http.Error(w, "Invalid expiration claim", http.StatusUnauthorized)
            return
        }

        if time.Now().Unix() > int64(exp) {
            http.Error(w, "Token has expired", http.StatusUnauthorized)
            return
        }

        // Token is valid and not expired; proceed with the request
        next.ServeHTTP(w, r)
    })
}
