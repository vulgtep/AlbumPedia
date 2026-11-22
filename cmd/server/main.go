package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"albumPedia.com/albumPedia/DB/sqlc"
	"albumPedia.com/albumPedia/pkg/handlers"
	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// ensureSchema garantiza que la tabla users y la columna user_id existan incluso si el volumen Docker fue creado en TPs anteriores.
func ensureSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
		`ALTER TABLE albums ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// 1. Establecer y Gestionar la Conexión mediante variables de entorno
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", getEnv("POSTGRES_USER", "postgres"))
	dbPassword := getEnv("DB_PASSWORD", getEnv("POSTGRES_PASSWORD", ""))
	dbName := getEnv("DB_NAME", getEnv("POSTGRES_DB", "albumpedia"))
	appPort := getEnv("APP_PORT", "8080")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error conectando a la base de datos: ", err)
	}
	defer db.Close()

	// 2. Verificar que la base de datos responde correctamente
	if err := db.Ping(); err != nil {
		log.Fatal("Error contactando a la base de datos: ", err)
	}

	if err := ensureSchema(db); err != nil {
		log.Fatal("Error asegurando esquema de usuarios: ", err)
	}

	// 3. Configurar el pool de conexiones para un acceso eficiente a la base de datos
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 4. Crear instancia del repositorio generado por sqlc
	repo := sqlc.New(db)

	// 5. Inyectar el repositorio en el servidor y registrar rutas
	server := handlers.NewServer(repo)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", server.HomeHandler)
	http.HandleFunc("/login", server.LoginHandler)
	http.HandleFunc("/register", server.RegisterHandler)
	http.HandleFunc("/logout", server.LogoutHandler)
	http.HandleFunc("/users/search", server.SearchUsersHandler)
	http.HandleFunc("/u/", server.UserProfileHandler)
	http.HandleFunc("/albums", server.CreateAlbumHandler)
	http.HandleFunc("/albums/", server.DeleteAlbumHandler)

	// 6. Iniciar el servidor HTTP
	fmt.Printf("Servidor escuchando en http://localhost:%s\n", appPort)
	if err := http.ListenAndServe(":"+appPort, nil); err != nil {
		log.Fatal("Error en el servidor HTTP: ", err)
	}
}
