package server

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = os.Getenv("JWT_KEY")

type AccountDetails struct {
    Name string `json:"name"`
}

func (s *Server) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
    var details AccountDetails
    if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Define a JWT claim structure
    claims := jwt.MapClaims{
        "name":  details.Name,
        "exp":   time.Now().Add(time.Hour * 72).Unix(), // token expires in 72 hours
    }

    // Generate the JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(jwtKey)) // replace with a secure secret key
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    // Respond with the generated token
    response := map[string]string{
        "token": tokenString,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}