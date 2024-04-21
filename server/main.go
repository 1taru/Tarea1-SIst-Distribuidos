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

type CacheServer struct {
	pb.UnimplementedCacheServiceServer
}

func (s *CacheServer) SetInCache(ctx context.Context, req *pb.SetInCacheRequest) (*pb.SetInCacheResponse, error) {
	key := req.GetKey()

	// Abre la conexión a la base de datos
	db, err := sql.Open("postgres", "user=postgres password=123 dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Realiza la consulta SQL para obtener el valor correspondiente a la clave
	var value string
	err = db.QueryRow("SELECT nomb_carrera FROM public.Universitario WHERE nomb_carrera = $1;", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("key '%s' not found in the database", key)
			return nil, fmt.Errorf("key '%s' not found in the database", key)
		}
		log.Fatalf("failed to query database: %v", err)
	}

	log.Printf("found value '%s' for key '%s'", value, key)
	return &pb.SetInCacheResponse{}, nil

}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCacheServiceServer(s, &CacheServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
