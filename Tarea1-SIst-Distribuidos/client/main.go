package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	pb "grpc_cache/proto"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr     = flag.String("addr", "localhost:50051", "the address to connect to")
	message  = ""
	aciertos = 0
	total    = 0
)

func main() {
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("---------------------------------------------------------------\n")
		fmt.Print("Ingrese búsqueda o 'exit' para salir: ")
		scanner.Scan()
		message = scanner.Text()

		if message == "exit" {
			return
		}

		total++

		valueInCache, err := HandleGetFromCache(*addr, message)
		if err != nil {
			log.Println(err)
		}

		if valueInCache == "" {
			start := time.Now()

			log.Println("No existe en la caché")

			success, err := HandleSetInCache(*addr, message, "value")
			if err != nil {
				log.Println(err)
			}

			if success {
				log.Println("Valor almacenado en la caché")
			}

			elapsed := time.Since(start)
			fmt.Printf("\nTiempo de almacenamiento en caché: %s \n", elapsed)

		} else {
			aciertos++
			start := time.Now()

			log.Println("Encontrado en la caché")
			log.Println("Valor en la caché:", valueInCache)

			elapsed := time.Since(start)
			fmt.Printf("\nTiempo de búsqueda en la caché: %s \n", elapsed)
		}

		fmt.Println("\nPorcentaje de aciertos: ", (aciertos*100)/total, "%")

	}

}
func HandleGetFromCache(addr string, key string) (string, error) {
	// Establece la conexión con el servidor gRPC
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("no se pudo conectar al servidor gRPC: %v", err)
	}
	defer conn.Close()

	// Crea un cliente gRPC
	c := pb.NewCacheServiceClient(conn)

	// Contexto con tiempo de espera
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Realiza la llamada al servidor gRPC
	response, err := c.GetFromCache(ctx, &pb.GetFromCacheRequest{Key: key})
	if err != nil {
		return "", fmt.Errorf("error al llamar al servidor gRPC: %v", err)
	}

	return response.Value, nil
}

func HandleSetInCache(addr string, key string, value string) (bool, error) {
	// Establece la conexión con el servidor gRPC
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, fmt.Errorf("no se pudo conectar al servidor gRPC: %v", err)
	}
	defer conn.Close()

	// Crea un cliente gRPC
	c := pb.NewCacheServiceClient(conn)

	// Contexto con tiempo de espera
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Realiza la llamada al servidor gRPC
	response, err := c.SetInCache(ctx, &pb.SetInCacheRequest{Key: key, Value: value})
	if err != nil {
		return false, fmt.Errorf("error al llamar al servidor gRPC: %v", err)
	}

	return response.Success, nil
}
