package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"

	pb "grpc_cache/proto"

	"google.golang.org/grpc"

	_ "github.com/lib/pq" // Importa el controlador de PostgreSQL
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedDatabaseServiceServer
	pb.UnimplementedCacheServiceServer
}

func (s *server) GetFromDatabase(ctx context.Context, req *pb.GetFromDatabaseRequest) (*pb.GetFromDatabaseResponse, error) {
	key := req.GetKey()

	// Abre la conexión a la base de datos
	db, err := sql.Open("postgres", "user=postgres password=123 dbname=mydatabase sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Realiza la consulta SQL para obtener el valor correspondiente a la clave
	var value string
	err = db.QueryRow("SELECT * FROM tabla2 WHERE nombre = $1", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("key '%s' not found in the database", key)
			return nil, fmt.Errorf("key '%s' not found in the database", key)
		}
		log.Fatalf("failed to query database: %v", err)
	}

	log.Printf("found value '%s' for key '%s'", value, key)
	return &pb.GetFromDatabaseResponse{
		Value: value,
	}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDatabaseServiceServer(s, &server{})
	pb.RegisterCacheServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
