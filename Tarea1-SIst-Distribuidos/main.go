package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // Importa el controlador de PostgreSQL
)

func main() {
	// Define la cadena de conexión para PostgreSQL
	connectionString := "postgresql://postgres:123@localhost/postgres?sslmode=disable"

	// Abre una nueva conexión a PostgreSQL
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("si se ingreso papu", err)
	}
	defer db.Close()

	// Intenta ejecutar una consulta de prueba
	rows, err := db.Query("SELECT 1")
	if err != nil {
		log.Fatal("Error al ejecutar la consulta de prueba:", err)
	}
	defer rows.Close()

	log.Println("¡Conexión exitosa a la base de datos PostgreSQL!")
}
