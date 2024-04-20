package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	port = flag.Int("port", 50051, "The server port")
	db   *sql.DB
	err  error
)

func main() {
	flag.Parse()

	// Establece la conexión con la base de datos PostgreSQL
	db, err = sql.Open("postgres", "user=postgres password=123 dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Conectado a la base de datos PostgreSQL")
	// Registra los manejadores de rutas HTTP
	http.HandleFunc("/getValue", getValueHandler)

	// Inicia el servidor HTTP
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server listening at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getValueHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	var value string
	err := db.QueryRow("SELECT apellido FROM tabla2 WHERE nombre = $1;", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("id '%s' not found in the database", key), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to query database: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"value": "%s"}`, value)
}
