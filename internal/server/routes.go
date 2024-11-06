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

    // Responds with some data about the application like open connections and such (Get only)
    r.HandleFunc("/health", s.healthHandler)

    // Post takes Username and Password -> validates password -> responds with a JWT token that holds basic jwt values + role of user and username
    // Get takes Username and Password -> validates password -> responds with the corresponding row in in Users Table without the password
    r.HandleFunc("/account", s.accountHandler)

    // Post takes Username, Role_Name and Password -> validates password -> responds with Status -> Assigns Role to User
    // Get takes Username and Password -> validates password -> responds with list of roles that are assigned to the user
    r.HandleFunc("/roles", s.rolesHandler)

    // Takes Username, Email and Pasword and puts them in the Users Table
    r.HandleFunc("/account_register", s.AccountRegisterHandlerDB).Methods(http.MethodPost)

    // Takes a role_name and writes it in the Roles Table
    r.HandleFunc("/roles_register", s.RolesRegisterHandlerDB).Methods(http.MethodPost)

    return r
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