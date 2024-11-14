package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UsernameStruct struct {
    Username string `json:"username"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

// HandleAccountJwt generates both an access token and a refresh token for the user
func (s *Server) HandleAccountJwt(w http.ResponseWriter, r *http.Request) {
    var usernameStruct UsernameStruct
    if err := json.NewDecoder(r.Body).Decode(&usernameStruct); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Check if the user exists in the database
    roles, err := s.db.GetRolesByUsername(usernameStruct.Username)
    if err != nil {
        http.Error(w, "Error querying database", http.StatusInternalServerError)
        return
    }
    if roles == nil {
        http.Error(w, "Invalid username", http.StatusUnauthorized)
        return
    }

    // Generate Access Token (short-lived)
    accessClaims := jwt.MapClaims{
        "username":  usernameStruct.Username,
        "role":      roles, 
        "exp":       time.Now().Add(1 * time.Minute).Unix(), // Access token expires in 1 min
        "iat":       time.Now().Unix(),
    }
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString([]byte(jwtKey))
    if err != nil {
        http.Error(w, "Error generating access token", http.StatusInternalServerError)
        return
    }

    // Generate Refresh Token (long-lived)
    refreshClaims := jwt.MapClaims{
        "username": usernameStruct.Username,
        "exp":      time.Now().Add(7 * 24 * time.Hour).Unix(), // Refresh token expires in 7 days
        "iat":      time.Now().Unix(),
    }
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString([]byte(jwtKey))
    if err != nil {
        http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
        return
    }

    response := TokenResponse{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}


func (s *Server) RefreshHandler(w http.ResponseWriter, r *http.Request) {
    var tokenReq struct {
        RefreshToken string `json:"refresh_token"`
    }

    // Decode the incoming request to get the refresh token
    if err := json.NewDecoder(r.Body).Decode(&tokenReq); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Parse the JWT token
    token, err := jwt.Parse(tokenReq.RefreshToken, func(token *jwt.Token) (interface{}, error) {
        return []byte(jwtKey), nil
    })
    if err != nil || !token.Valid {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }

    // Extract claims from the token and ensure they are in the correct format
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || claims["username"] == nil {
        http.Error(w, "Invalid token claims", http.StatusUnauthorized)
        return
    }

    // Extract the username from claims
    username := claims["username"].(string)

    // Check if the user exists in the database
    roles, err := s.db.GetRolesByUsername(username)
    if err != nil {
        http.Error(w, "Error querying database", http.StatusInternalServerError)
        return
    }
    if roles == nil {
        http.Error(w, "Invalid username", http.StatusUnauthorized)
        return
    }

    // Generate Access Token (short-lived)
    accessClaims := jwt.MapClaims{
        "username":  username,
        "role":      roles, 
        "exp":       time.Now().Add(1 * time.Minute).Unix(), // Access token expires in 1 min
        "iat":       time.Now().Unix(),
    }

    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString([]byte(jwtKey))
    if err != nil {
        http.Error(w, "Error generating access token", http.StatusInternalServerError)
        return
    }

    response := map[string]string{
        "token": accessTokenString,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}


func (s *Server) HandleAccountDB(w http.ResponseWriter, r *http.Request) {
    var usernameStruct UsernameStruct
    if err := json.NewDecoder(r.Body).Decode(&usernameStruct); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Check if the user exists in the database
    user, err := s.db.GetUserByUsername(usernameStruct.Username)
    if err != nil {
        http.Error(w, "Error querying database", http.StatusInternalServerError)
        return
    }
    if user == nil {
        http.Error(w, "Invalid username", http.StatusUnauthorized)
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

type AccountRegisterRequest struct {
    Username string `json:"username"`
    Email string `json:"email"`
    Password string `json:"password"`
}

//Takes the AccountRegisterRequest struct params and writes them in the Users Table to register an new User, ID and created_at get filled automaticaly
func (s *Server) AccountRegisterHandlerDB(w http.ResponseWriter, r *http.Request) {
    var req AccountRegisterRequest

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

type RolesRegisterRequest struct {
    Role_Name string `json:"role_name"`
}

//Takes the RolesRegisterRequest struct params and writes them in the Roles Table to register an new Role, ID get's filled automaticaly
func (s *Server) RolesRegisterHandlerDB(w http.ResponseWriter, r *http.Request) {
    var req RolesRegisterRequest

    // Parse the JSON request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Insert the user into the database
    err := s.db.CreateRole(req.Role_Name)
    if err != nil {
        http.Error(w, "Failed to register role", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Role registered successfully"))
}

type RolesAssignment struct {
    Username string `json:"username"`
    Role_Name string `json:"role_name"`
}

func (s *Server) HandleRoleAddedToUserDB(w http.ResponseWriter, r *http.Request) {
    var req RolesAssignment

    // Parse the JSON request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Insert the user into the database
    err := s.db.AssignRoleToUser(req.Username, req.Role_Name)
    if err != nil {
        http.Error(w, "Failed to assign role", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Role assigned successfully"))
}

func (s *Server) HandleGetUserRole(w http.ResponseWriter, r *http.Request) {
    var usernameStruct UsernameStruct
    if err := json.NewDecoder(r.Body).Decode(&usernameStruct); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Check if the user exists in the database
    roles, err := s.db.GetRolesByUsername(usernameStruct.Username)
    if err != nil {
        http.Error(w, "Error querying database", http.StatusInternalServerError)
        return
    }
    if roles == nil {
        http.Error(w, "Invalid username", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(roles)
}