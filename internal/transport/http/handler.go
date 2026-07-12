package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"usersvc/internal/repository"
	"usersvc/internal/service"
)

type Handler struct {
	service *service.UserService
}

func NewHandler(service *service.UserService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users", h.ListUsers)
	mux.HandleFunc("GET /users/{id}", h.GetUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)
	mux.HandleFunc("GET /health", h.Health)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid req body", http.StatusBadRequest)
		return
	}
	usr, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		slog.Error("couldnt reg the user", "err", err)
		http.Error(w, "couldn't reg the user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(usr)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		slog.Error("couldnt list users", "error", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missign user id", http.StatusInternalServerError)
		return
	}
	usr, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "couldnt find any such user", http.StatusNotFound)
			return
		}
		slog.Error("couldn't get user", "error", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(usr)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "couldnt find any such user", http.StatusNotFound)
			return
		}
		slog.Error("couldnt delete the user", "error", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusNoContent)
}
