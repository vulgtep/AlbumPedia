package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"albumPedia.com/albumPedia/DB/sqlc"
	"albumPedia.com/albumPedia/views"
)

// Server contiene el repositorio generado por sqlc
type Server struct {
	Queries *sqlc.Queries
}

func NewServer(queries *sqlc.Queries) *Server {
	return &Server{Queries: queries}
}

// GET / - Obtiene los álbumes del usuario autenticado y renderiza su perfil con formulario
func (s *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, ok := s.getAuthenticatedUser(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	albums, err := s.Queries.ListAlbumsByUser(r.Context(), currentUser.ID)
	if err != nil {
		http.Error(w, "Error obteniendo álbumes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := views.HomePage(currentUser, *currentUser, albums, true).Render(r.Context(), w); err != nil {
		http.Error(w, "Error renderizando la página", http.StatusInternalServerError)
	}
}

// POST /albums - Crea el álbum para el usuario autenticado y devuelve el fragmento EntityList para HTMX
func (s *Server) CreateAlbumHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, ok := s.getAuthenticatedUser(r)
	if !ok {
		http.Error(w, "Sesión expirada. Iniciá sesión nuevamente.", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error procesando el formulario: "+err.Error(), http.StatusBadRequest)
		return
	}

	titulo := strings.TrimSpace(r.FormValue("titulo"))
	banda := strings.TrimSpace(r.FormValue("banda"))
	genero := strings.TrimSpace(r.FormValue("genero"))

	if titulo == "" || banda == "" || genero == "" {
		http.Error(w, "Título, banda y género son campos obligatorios", http.StatusBadRequest)
		return
	}

	arg := sqlc.CreateAlbumParams{
		UserID:          currentUser.ID,
		Titulo:          titulo,
		Banda:           banda,
		Genero:          genero,
		Productor:       parseNullString(r.FormValue("productor")),
		ListaArtistas:   parseNullString(r.FormValue("lista_artistas")),
		Portada:         parseNullString(r.FormValue("portada")),
		Contraportada:   parseNullString(r.FormValue("contraportada")),
		DuracionMinutos: parseNullInt32(r.FormValue("duracion_minutos")),
		FechaSalida:     parseNullTime(r.FormValue("fecha_salida")),
		LinkEscucha:     parseNullString(r.FormValue("link_escucha")),
		Observacion:     parseNullString(r.FormValue("observacion")),
	}

	if _, err := s.Queries.CreateAlbum(r.Context(), arg); err != nil {
		http.Error(w, "Error creando álbum: "+err.Error(), http.StatusInternalServerError)
		return
	}

	albums, err := s.Queries.ListAlbumsByUser(r.Context(), currentUser.ID)
	if err != nil {
		http.Error(w, "Error obteniendo álbumes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := views.EntityList(albums, true, currentUser.Username).Render(r.Context(), w); err != nil {
		http.Error(w, "Error renderizando la lista de álbumes", http.StatusInternalServerError)
	}
}

// Handler para /albums/{id}, /albums/{id}/edit y /albums/{id}/card
// Verifica siempre que el usuario autenticado sea el dueño del álbum (WHERE id = $1 AND user_id = $2).
func (s *Server) DeleteAlbumHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := s.getAuthenticatedUser(r)
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/albums/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "ID de álbum inválido", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "ID de álbum inválido", http.StatusBadRequest)
		return
	}

	// 1. GET /albums/{id}/edit -> Devuelve el formulario de edición inline si es el propietario
	if r.Method == http.MethodGet && len(parts) == 2 && parts[1] == "edit" {
		album, err := s.Queries.GetAlbumByOwner(r.Context(), sqlc.GetAlbumByOwnerParams{
			ID:     int32(id),
			UserID: currentUser.ID,
		})
		if err != nil {
			http.Error(w, "No tenés permiso para editar este álbum", http.StatusForbidden)
			return
		}
		_ = views.AlbumEditCard(album).Render(r.Context(), w)
		return
	}

	// 2. GET /albums/{id}/card -> Cancela la edición y devuelve la tarjeta original del propietario
	if r.Method == http.MethodGet && len(parts) == 2 && parts[1] == "card" {
		album, err := s.Queries.GetAlbumByOwner(r.Context(), sqlc.GetAlbumByOwnerParams{
			ID:     int32(id),
			UserID: currentUser.ID,
		})
		if err != nil {
			http.Error(w, "No autorizado", http.StatusForbidden)
			return
		}
		_ = views.AlbumCard(album, true).Render(r.Context(), w)
		return
	}

	// 3. PUT /albums/{id} -> Actualiza el álbum únicamente si pertenece al usuario autenticado
	if r.Method == http.MethodPut && len(parts) == 1 {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error procesando el formulario", http.StatusBadRequest)
			return
		}

		titulo := strings.TrimSpace(r.FormValue("titulo"))
		banda := strings.TrimSpace(r.FormValue("banda"))
		genero := strings.TrimSpace(r.FormValue("genero"))
		if titulo == "" || banda == "" || genero == "" {
			http.Error(w, "Título, banda y género son obligatorios", http.StatusBadRequest)
			return
		}

		updatedAlbum, err := s.Queries.UpdateAlbumByOwner(r.Context(), sqlc.UpdateAlbumByOwnerParams{
			ID:              int32(id),
			UserID:          currentUser.ID,
			Titulo:          titulo,
			Banda:           banda,
			Genero:          genero,
			Productor:       parseNullString(r.FormValue("productor")),
			ListaArtistas:   parseNullString(r.FormValue("lista_artistas")),
			Portada:         parseNullString(r.FormValue("portada")),
			Contraportada:   parseNullString(r.FormValue("contraportada")),
			DuracionMinutos: parseNullInt32(r.FormValue("duracion_minutos")),
			FechaSalida:     parseNullTime(r.FormValue("fecha_salida")),
			LinkEscucha:     parseNullString(r.FormValue("link_escucha")),
			Observacion:     parseNullString(r.FormValue("observacion")),
		})
		if err != nil {
			http.Error(w, "No tenés permiso para modificar este álbum", http.StatusForbidden)
			return
		}

		_ = views.AlbumCard(updatedAlbum, true).Render(r.Context(), w)
		return
	}

	// 4. DELETE /albums/{id} -> Elimina el álbum únicamente si pertenece al usuario autenticado
	if r.Method == http.MethodDelete && len(parts) == 1 {
		err = s.Queries.DeleteAlbumByOwner(r.Context(), sqlc.DeleteAlbumByOwnerParams{
			ID:     int32(id),
			UserID: currentUser.ID,
		})
		if err != nil {
			http.Error(w, "Error eliminando álbum", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func parseNullString(val string) sql.NullString {
	trimmed := strings.TrimSpace(val)
	return sql.NullString{
		String: trimmed,
		Valid:  trimmed != "",
	}
}

func parseNullInt32(val string) sql.NullInt32 {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return sql.NullInt32{Valid: false}
	}
	n, err := strconv.Atoi(trimmed)
	if err != nil || n <= 0 {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{
		Int32: int32(n),
		Valid: true,
	}
}

func parseNullTime(val string) sql.NullTime {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  t,
		Valid: true,
	}
}
