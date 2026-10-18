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

func main() {
	// 1. Establecer y Gestionar la Conexión
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:5432/albumpedia?sslmode=disable", dbHost)
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

	// 5. Inyectar el repositorio en el servidor
	server := handlers.NewServer(repo)

	http.HandleFunc("/albums", server.AlbumsHandlerSinID)
	http.HandleFunc("/albums/", server.AlbumsHandlerConID)

	// 6. Iniciar el servidor HTTP
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	fmt.Println("Servidor escuchando en http://localhost:8080")
	err2 := http.ListenAndServe(":8080", nil)
	if err2 != nil {
		fmt.Println("Error:", err2)
	}
}
