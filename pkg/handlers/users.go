package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"albumPedia.com/albumPedia/DB/sqlc"
	"albumPedia.com/albumPedia/pkg/auth"
	"albumPedia.com/albumPedia/views"
)

func (s *Server) getAuthenticatedUser(r *http.Request) (*sqlc.User, bool) {
	userID, ok := auth.GetSessionUserID(r)
	if !ok {
		return nil, false
	}
	user, err := s.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		return nil, false
	}
	return &user, true
}

// GET /login y POST /login
func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if _, ok := s.getAuthenticatedUser(r); ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		_ = views.LoginPage("").Render(r.Context(), w)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		_ = views.LoginPage("Error procesando el formulario").Render(r.Context(), w)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	user, err := s.Queries.GetUserByUsername(r.Context(), username)
	if err != nil || !auth.VerifyPassword(password, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = views.LoginPage("Usuario o contraseña incorrectos").Render(r.Context(), w)
		return
	}

	auth.SetSessionCookie(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GET /register y POST /register
func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if _, ok := s.getAuthenticatedUser(r); ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		_ = views.RegisterPage("").Render(r.Context(), w)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		_ = views.RegisterPage("Error procesando el formulario").Render(r.Context(), w)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	if len(username) < 3 || len(username) > 50 || strings.ContainsAny(username, " /?&#") {
		w.WriteHeader(http.StatusBadRequest)
		_ = views.RegisterPage("El nombre de usuario debe tener entre 3 y 50 caracteres sin espacios.").Render(r.Context(), w)
		return
	}
	if len(password) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		_ = views.RegisterPage("La contraseña debe tener al menos 4 caracteres.").Render(r.Context(), w)
		return
	}

	if _, err := s.Queries.GetUserByUsername(r.Context(), username); err == nil {
		w.WriteHeader(http.StatusConflict)
		_ = views.RegisterPage("Ese nombre de usuario ya está registrado.").Render(r.Context(), w)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "Error interno de seguridad", http.StatusInternalServerError)
		return
	}

	newUser, err := s.Queries.CreateUser(r.Context(), sqlc.CreateUserParams{
		Username:     username,
		PasswordHash: hash,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = views.RegisterPage("No se pudo crear el usuario.").Render(r.Context(), w)
		return
	}

	auth.SetSessionCookie(w, newUser.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// POST /logout
func (s *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// GET /users/search?q=... - Devuelve resultados parciales de búsqueda de usuarios para HTMX
func (s *Server) SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, ok := s.getAuthenticatedUser(r); !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		_ = views.UserSearchResults(nil, "").Render(r.Context(), w)
		return
	}

	users, err := s.Queries.SearchUsers(r.Context(), sql.NullString{String: q, Valid: true})
	if err != nil {
		http.Error(w, "Error buscando usuarios", http.StatusInternalServerError)
		return
	}

	_ = views.UserSearchResults(users, q).Render(r.Context(), w)
}

// GET /u/{username} - Renderiza el perfil de un usuario (sin formulario si es cuenta ajena)
func (s *Server) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, ok := s.getAuthenticatedUser(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	targetUsername := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/u/"))
	if targetUsername == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	profileUser, err := s.Queries.GetUserByUsername(r.Context(), targetUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Error cargando perfil", http.StatusInternalServerError)
		return
	}

	albums, err := s.Queries.ListAlbumsByUser(r.Context(), profileUser.ID)
	if err != nil {
		http.Error(w, "Error obteniendo álbumes del usuario", http.StatusInternalServerError)
		return
	}

	isOwner := currentUser.ID == profileUser.ID
	if err := views.HomePage(currentUser, profileUser, albums, isOwner).Render(r.Context(), w); err != nil {
		http.Error(w, "Error renderizando perfil", http.StatusInternalServerError)
	}
}
