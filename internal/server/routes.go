package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) RegisterRoutes() http.Handler {
    r := mux.NewRouter()

    r.HandleFunc("/", s.HelloWorldHandler)

    r.HandleFunc("/health", s.healthHandler)

    r.HandleFunc("/account", s.accountHandler)

    r.HandleFunc("/roles", s.rolesHandler)

    r.HandleFunc("/register", s.RegisterHandler).Methods(http.MethodPost)

    return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
    resp := make(map[string]string)
    resp["message"] = "Hello World"

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
        s.HandleRoleAddedDB(w, r)
    case http.MethodGet:
        s.HandleGetUserRole(w, r)
    default:
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
}