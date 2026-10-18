package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"albumPedia.com/albumPedia/DB/sqlc"
)

// Server contiene el repositorio generado por sqlc
type Server struct {
	Queries *sqlc.Queries
}

func NewServer(queries *sqlc.Queries) *Server {
	return &Server{Queries: queries}
}

// AlbumsHandler maneja la colección general (/albums)
func (s *Server) AlbumsHandlerSinID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listAlbums(w, r)
	case http.MethodPost:
		s.createAlbum(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// SingleAlbumHandler maneja un álbum específico (/albums/{id})
func (s *Server) AlbumsHandlerConID(w http.ResponseWriter, r *http.Request) {
	// ["", "albums", "ID"]
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "URL inválida", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "ID de álbum inválido", http.StatusBadRequest)
		return
	}

	albumID := int32(id)

	switch r.Method {
	case http.MethodGet:
		s.getAlbum(w, r, albumID)
	case http.MethodPut:
		s.updateAlbum(w, r, albumID)
	case http.MethodDelete:
		s.deleteAlbum(w, r, albumID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /albums
func (s *Server) createAlbum(w http.ResponseWriter, r *http.Request) {
	var arg sqlc.CreateAlbumParams
	if err := json.NewDecoder(r.Body).Decode(&arg); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(arg.Titulo) == "" {
		http.Error(w, "El título no puede estar vacío", http.StatusBadRequest)
		return
	}

	album, err := s.Queries.CreateAlbum(r.Context(), arg)
	if err != nil {
		http.Error(w, "Error creando álbum", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // HTTP 201
	json.NewEncoder(w).Encode(album)
}

// GET /albums
func (s *Server) listAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := s.Queries.ListAlbums(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo álbumes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(albums)
}

// GET /albums/{id}
func (s *Server) getAlbum(w http.ResponseWriter, r *http.Request, id int32) {
	album, err := s.Queries.GetAlbum(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Álbum no encontrado", http.StatusNotFound) // HTTP 404
			return
		}
		http.Error(w, "Error de base de datos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(album)
}

// PUT /albums/{id}
func (s *Server) updateAlbum(w http.ResponseWriter, r *http.Request, id int32) {
	var arg sqlc.UpdateAlbumParams
	if err := json.NewDecoder(r.Body).Decode(&arg); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	arg.ID = id
	if strings.TrimSpace(arg.Titulo) == "" {
		http.Error(w, "El título no puede estar vacío", http.StatusBadRequest)
		return
	}

	album, err := s.Queries.UpdateAlbum(r.Context(), arg)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Álbum no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error actualizando álbum", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(album)
}

// DELETE /albums/{id}
func (s *Server) deleteAlbum(w http.ResponseWriter, r *http.Request, id int32) {
	err := s.Queries.DeleteAlbum(r.Context(), id)
	if err != nil {
		http.Error(w, "Error eliminando álbum", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // HTTP 204
}
