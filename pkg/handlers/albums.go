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

// GET / - Obtiene todos los álbumes y renderiza la vista principal con templ
func (s *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	albums, err := s.Queries.ListAlbums(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo álbumes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := views.HomePage(albums).Render(r.Context(), w); err != nil {
		http.Error(w, "Error renderizando la página", http.StatusInternalServerError)
	}
}

// POST /albums - Crea el álbum en BD y devuelve el fragmento EntityList para HTMX
func (s *Server) CreateAlbumHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	albums, err := s.Queries.ListAlbums(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo álbumes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := views.EntityList(albums).Render(r.Context(), w); err != nil {
		http.Error(w, "Error renderizando la lista de álbumes", http.StatusInternalServerError)
	}
}

// DELETE /albums/{id} - Elimina un álbum por ID y devuelve 200 OK vacío para HTMX
func (s *Server) DeleteAlbumHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/albums/"))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de álbum inválido", http.StatusBadRequest)
		return
	}

	if err := s.Queries.DeleteAlbum(r.Context(), int32(id)); err != nil {
		http.Error(w, "Error eliminando álbum", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
