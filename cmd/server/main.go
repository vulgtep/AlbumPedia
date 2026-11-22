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

func main() {
	// 1. Establecer y Gestionar la Conexión mediante variables de entorno
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", getEnv("POSTGRES_USER", "postgres"))
	dbPassword := getEnv("DB_PASSWORD", getEnv("POSTGRES_PASSWORD", ""))
	dbName := getEnv("DB_NAME", getEnv("POSTGRES_DB", "albumpedia"))

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

	// 3. Configurar el pool de conexiones para un acceso eficiente a la base de datos
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 4. Crear instancia del repositorio generado por sqlc
	repo := sqlc.New(db)

	// 5. Inyectar el repositorio en el servidor y registrar rutas SSR
	server := handlers.NewServer(repo)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", server.HomeHandler)
	http.HandleFunc("/albums", server.CreateAlbumHandler)
	http.HandleFunc("/albums/", server.DeleteAlbumHandler)

	// 6. Iniciar el servidor HTTP
	fmt.Println("Servidor SSR escuchando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error en el servidor HTTP: ", err)
	}
}
