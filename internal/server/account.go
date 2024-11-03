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
    Username string `json:"username"`
    Password string `json:"password"`
}

func (s *Server) HandleAccountJwt(w http.ResponseWriter, r *http.Request) {
    var details AccountDetails
    if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    claims := jwt.MapClaims{
        "username":  details.Username,
        "exp":       time.Now().Add(time.Hour * 72).Unix(), // token expires in 72 hours
        "iat":       time.Now().Unix(),
        "role":      "not implemented",  //todo
    }

    // Generate the JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(jwtKey)) // replace with a secure secret key
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    response := map[string]string{
        "token": tokenString,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

type RegisterRequest struct {
    Username string `json:"username"`
    Email string `json:"email"`
    Password string `json:"password"`
}

func (s *Server) HandleAccountDB(w http.ResponseWriter, r *http.Request) {
	var details AccountDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check if the user exists in the database
	user, err := s.db.GetUserByUsernameAndPassword(details.Username, details.Password)
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Respond with the user's information (excluding password)
	response := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest

    // Parse the JSON request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Insert the user into the database
    err := s.db.CreateUser(req.Username, req.Email, req.Password)
    if err != nil {
        http.Error(w, "Failed to register user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("User registered successfully"))
}