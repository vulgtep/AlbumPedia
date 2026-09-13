package sqlc

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) (*Queries, *sql.DB) {
	// conectamos a la base de datos
	connString := "postgres://postgres:postgres@localhost:5432/albumpedia?sslmode=disable"
	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("Fallo al conectar a la DB de prueba: %v", err)
	}
	return New(db), db
}

// createRandomAlbum es una función auxiliar que inserta un álbum genérico.
// Esto nos evita repetir código en los tests que necesitan que ya exista un álbum (como Get, Update o Delete).
func createRandomAlbum(t *testing.T, queries *Queries) Album {
	ctx := context.Background()
	arg := CreateAlbumParams{
		Titulo:          "Álbum Genérico para Tests",
		Genero:          "Rock Alternativo",
		Banda:           "Banda de Prueba",
		Productor:       sql.NullString{String: "Productor Test", Valid: true},
		ListaArtistas:   sql.NullString{String: "Artista 1, Artista 2", Valid: true},
		Portada:         sql.NullString{String: "portada.jpg", Valid: true},
		Contraportada:   sql.NullString{String: "contraportada.jpg", Valid: true},
		DuracionMinutos: sql.NullInt32{Int32: 45, Valid: true},
		FechaSalida:     sql.NullTime{Time: time.Now(), Valid: true},
		LinkEscucha:     sql.NullString{String: "https://link.test", Valid: true},
		Observacion:     sql.NullString{String: "Observación de prueba", Valid: true},
	}

	album, err := queries.CreateAlbum(ctx, arg)
	if err != nil {
		t.Fatalf("Error creando álbum auxiliar: %v", err)
	}
	return album
}

func TestCreateAlbum(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close() // Cerramos la conexión al terminar el test

	ctx := context.Background()
	arg := CreateAlbumParams{
		Titulo:          "The Dark Side of the Moon",
		Genero:          "Progressive Rock",
		Banda:           "Pink Floyd",
		Productor:       sql.NullString{String: "Alan Parsons", Valid: true},
		ListaArtistas:   sql.NullString{Valid: false}, // Campo Nulo
		Portada:         sql.NullString{Valid: false},
		Contraportada:   sql.NullString{Valid: false},
		DuracionMinutos: sql.NullInt32{Int32: 43, Valid: true},
		FechaSalida:     sql.NullTime{Valid: false},
		LinkEscucha:     sql.NullString{Valid: false},
		Observacion:     sql.NullString{Valid: false},
	}

	album, err := queries.CreateAlbum(ctx, arg)
	if err != nil {
		t.Fatalf("Error creando álbum: %v", err)
	}

	// Validaciones
	if album.ID == 0 {
		t.Errorf("Se esperaba un ID mayor a 0, pero se obtuvo %d", album.ID)
	}
	if album.Titulo != arg.Titulo {
		t.Errorf("Se esperaba título '%s', pero se obtuvo '%s'", arg.Titulo, album.Titulo)
	}
	if album.Banda != arg.Banda {
		t.Errorf("Se esperaba banda '%s', pero se obtuvo '%s'", arg.Banda, album.Banda)
	}
}

func TestGetAlbum(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// 1. Insertamos un álbum en la base de datos
	nuevoAlbum := createRandomAlbum(t, queries)

	// 2. Lo buscamos por su ID
	albumEncontrado, err := queries.GetAlbum(ctx, nuevoAlbum.ID)
	if err != nil {
		t.Fatalf("Error obteniendo álbum: %v", err)
	}

	// 3. Comparamos que sea el mismo
	if albumEncontrado.ID != nuevoAlbum.ID {
		t.Errorf("Se esperaba ID %d, pero se obtuvo %d", nuevoAlbum.ID, albumEncontrado.ID)
	}
	if albumEncontrado.Titulo != nuevoAlbum.Titulo {
		t.Errorf("Se esperaba Título '%s', pero se obtuvo '%s'", nuevoAlbum.Titulo, albumEncontrado.Titulo)
	}
}

func TestListAlbums(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// 1. Aseguramos que haya al menos dos álbumes en la tabla
	createRandomAlbum(t, queries)
	createRandomAlbum(t, queries)

	// 2. Obtenemos la lista de álbumes
	albums, err := queries.ListAlbums(ctx)
	if err != nil {
		t.Fatalf("Error listando álbumes: %v", err)
	}

	// 3. Verificamos la cantidad
	if len(albums) < 2 {
		t.Errorf("Se esperaban al menos 2 álbumes, se obtuvieron %d", len(albums))
	}
}

func TestUpdateAlbum(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// 1. Creamos el álbum original
	albumOriginal := createRandomAlbum(t, queries)

	// 2. Preparamos los parámetros, modificando solo un par de campos
	arg := UpdateAlbumParams{
		ID:              albumOriginal.ID,
		Titulo:          "Título Modificado",
		Genero:          "Pop",
		Banda:           albumOriginal.Banda,
		Productor:       albumOriginal.Productor,
		ListaArtistas:   albumOriginal.ListaArtistas,
		Portada:         albumOriginal.Portada,
		Contraportada:   albumOriginal.Contraportada,
		DuracionMinutos: sql.NullInt32{Int32: 50, Valid: true}, // Cambio
		FechaSalida:     albumOriginal.FechaSalida,
		LinkEscucha:     albumOriginal.LinkEscucha,
		Observacion:     albumOriginal.Observacion,
	}

	// 3. Ejecutamos la actualización
	albumActualizado, err := queries.UpdateAlbum(ctx, arg)
	if err != nil {
		t.Fatalf("Error actualizando álbum: %v", err)
	}

	// 4. Verificamos que los campos se hayan actualizado
	if albumActualizado.Titulo != "Título Modificado" {
		t.Errorf("Se esperaba el título actualizado, pero se obtuvo '%s'", albumActualizado.Titulo)
	}
	if albumActualizado.Genero != "Pop" {
		t.Errorf("Se esperaba el género actualizado, pero se obtuvo '%s'", albumActualizado.Genero)
	}
}

func TestDeleteAlbum(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// 1. Creamos un álbum
	album := createRandomAlbum(t, queries)

	// 2. Lo borramos
	err := queries.DeleteAlbum(ctx, album.ID)
	if err != nil {
		t.Fatalf("Error borrando álbum: %v", err)
	}

	// 3. Intentamos buscarlo de nuevo
	_, err = queries.GetAlbum(ctx, album.ID)

	// 4. Debería darnos un error porque ya no existe (sql.ErrNoRows)
	if err != sql.ErrNoRows {
		t.Errorf("Se esperaba error sql.ErrNoRows (álbum no encontrado), pero se obtuvo: %v", err)
	}
}
