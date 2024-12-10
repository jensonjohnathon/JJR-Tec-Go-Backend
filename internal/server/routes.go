package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) RegisterRoutes() http.Handler {
    r := mux.NewRouter()

    // Just a default route with a welcome message (Get only)
    r.HandleFunc("/", s.defaultRouteHandler)

    // takes Username and Password -> validates password -> responds with a JWT token that holds basic jwt values + role of user and username
    // Protected by Password check
    r.HandleFunc("/account", s.accountHandler).Methods(http.MethodPost)

    // For JWT Refresh Tokens
    r.HandleFunc("/refresh", s.RefreshHandler)

    // Define protected routes with middleware
    s.registerProtectedRoutes(r)

    return s.CORS(r)
}

// registerProtectedRoutes sets up the protected routes under "/protected" with authentication middleware applied
func (s *Server) registerProtectedRoutes(r *mux.Router) {
    protected := r.PathPrefix("/protected").Subrouter()
    protected.Use(s.AuthMiddleware) // Apply authentication middleware to all /protected routes

    // POST takes username, email and password and registers a User with that data
    protected.HandleFunc("/account_register", s.AccountRegisterHandlerDB).Methods(http.MethodPost)

    // takes Username and Password -> validates password -> responds with the corresponding row in in Users Table without the password
    protected.HandleFunc("/account", s.accountHandler).Methods(http.MethodGet)

    // Post takes Username, Role_Name and Password -> validates password -> responds with Status -> Assigns Role to User
    // Get takes Username and Password -> validates password -> responds with list of roles that are assigned to the user
    protected.HandleFunc("/roles", s.rolesHandler)

    // Takes a role_name and writes it in the Roles Table
    protected.HandleFunc("/roles_register", s.RolesRegisterHandlerDB).Methods(http.MethodPost)

    // Responds with some data about the application like open connections and such (Get only)
    protected.HandleFunc("/health", s.healthHandler)
}

func (s *Server) CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent) // No content for OPTIONS requests
            return
        }

        next.ServeHTTP(w, r)
    })
}

func (s *Server) defaultRouteHandler(w http.ResponseWriter, r *http.Request) {
    resp := make(map[string]string)
    resp["message"] = "JJR Backend"

    jsonResp, err := json.Marshal(resp)
    if err != nil {
        log.Fatalf("error handling JSON marshal. Err: %v", err)
    }

    _, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    jsonResp, err := json.Marshal(s.db.Health())

    if err != nil {
        log.Fatalf("error handling JSON marshal. Err: %v", err)
    }

    _, _ = w.Write(jsonResp)
}

func (s *Server) accountHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        s.HandleAccountJwt(w, r)
    case http.MethodGet:
        s.HandleAccountDB(w, r)
    default:
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
}

func (s *Server) rolesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        s.HandleRoleAddedToUserDB(w, r)
    case http.MethodGet:
        s.HandleGetUserRole(w, r)
    default:
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
}