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

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr     = flag.String("addr", "localhost:50051", "the address to connect to")
	message  = ""
	aciertos = 0
	total    = 0
)

type CacheClient struct {
	cc pb.CacheServiceClient
}

func (c *CacheClient) GetFromCache(ctx context.Context, req *pb.GetFromCacheRequest) (*pb.GetFromCacheResponse, error) {
	// Conexión a Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	// Verificación de conexión con Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar al servidor Redis: %v", err)
	}

	// Obtención del valor desde Redis
	value, err := redisClient.Get(ctx, req.Key).Result()
	if err != nil {
		// Maneja el caso en el que la clave no existe en la caché
		if err == redis.Nil {
			log.Printf("clave '%s' no encontrada en la caché", req.Key)
			return nil, nil // Retorna nil para indicar que no se encontró la clave
		}
		log.Fatalf("error al obtener el valor de la caché: %v", err)
	}

	return &pb.GetFromCacheResponse{
		Value: value,
	}, nil
}

func (c *CacheClient) SetInCache(ctx context.Context, req *pb.SetInCacheRequest) (*pb.SetInCacheResponse, error) {
	// Primero, envía la solicitud al servidor gRPC para establecer el valor en la caché
	response, err := c.cc.SetInCache(ctx, req)
	if err != nil {
		return nil, err
	}
	// Si la operación en el servidor gRPC fue exitosa, ingresa el valor en Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	// Verifica la conexión con el servidor Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar al servidor Redis: %v", err)
	}

	// Ingresa el valor en Redis
	if err := redisClient.Set(ctx, req.Key, req.Value, 0).Err(); err != nil {
		log.Fatalf("error al ingresar el valor en Redis: %v", err)
	}

	// Retorna la respuesta del servidor gRPC
	return response, nil
}

/*
	func (c *client) GetFromCache(ctx context.Context, req *pb.GetFromCacheRequest) (*pb.GetFromCacheResponse, error) {
		redisClient := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
		defer redisClient.Close()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Fatalf("no se pudo conectar al servidor Redis: %v", err)
		}
		_, err := redisClient.Ping(context.Background()).Result()
		log.Printf("Conectándose al servidor gRPC en %s", *addr)
		if err != nil {
			log.Fatal(err)
		}
		value, err := redisClient.Get(ctx, req.Key).Result()
		if err != nil {
			// Maneja el caso en el que la clave no existe en la caché
			if err == redis.Nil {
				log.Printf("clave '%s' no encontrada en la caché", req.Key)
				return nil, nil // Retorna nil para indicar que no se encontró la clave
			}
			log.Fatalf("error al obtener el valor de la caché: %v", err)
		}
		return &pb.GetFromCacheResponse{
			Value: value,
		}, nil

}

	func (c *client) SetInCache(ctx context.Context, req *pb.SetInCacheRequest) (*pb.SetInCacheResponse, error) {
		// Primero, envía la solicitud al servidor gRPC para establecer el valor en la caché
		response, err := c.SetInCache(ctx, req)
		if err != nil {
			return nil, err
		}

		// Si la operación en el servidor gRPC fue exitosa, ingresa el valor en Redis
		redisClient := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		// Cierra el cliente Redis al final de la función
		defer redisClient.Close()

		// Verifica la conexión con el servidor Redis
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Fatalf("no se pudo conectar al servidor Redis: %v", err)
		}

		// Ingresa el valor en Redis
		if err := redisClient.Set(ctx, req.Key, req.Value, 0).Err(); err != nil {
			log.Fatalf("error al ingresar el valor en Redis: %v", err)
		}

		// Retorna la respuesta del servidor gRPC
		return response, nil
	}
*/
func main() {
	flag.Parse()

	/*
		redisClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{":16379", ":16380", ":16381", ":16382", ":16383", ":16384"},

			// To route commands by latency or randomly, enable one of the following.
			//RouteByLatency: true,
			//RouteRandomly: true,
		})
	*/
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("no se pudo conectar al servidor gRPC: %v", err)
	}
	defer conn.Close()

	client := pb.NewCacheServiceClient(conn)
	cacheClient := &CacheClient{cc: client}
	ctx := context.Background()

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

		// Crea una solicitud para obtener el valor de la caché
		getRequest := &pb.GetFromCacheRequest{
			Key: message,
		}

		// Llama a la función GetFromCache para obtener el valor de la caché
		ValueInCache, err := cacheClient.GetFromCache(ctx, getRequest)
		if err != nil {
			log.Println(err)
		}

		if ValueInCache == nil || ValueInCache.Value == "" {
			start := time.Now()

			log.Println("No existe en la caché")

			// Crea una solicitud para establecer el valor en la caché
			setRequest := &pb.SetInCacheRequest{
				Key:   message,
				Value: ValueInCache.GetValue(),
			}

			// Llama a la función SetInCache para almacenar el valor en la caché
			_, err := cacheClient.SetInCache(ctx, setRequest)
			if err != nil {
				log.Printf("error al llamar a SetInCache: %v", err)
				return
			}

			log.Println("Valor almacenado en la caché")

			elapsed := time.Since(start)
			fmt.Printf("\nTiempo de almacenamiento en caché: %s \n", elapsed)

		} else {
			aciertos++
			start := time.Now()

			log.Println("Encontrado en la caché")
			log.Println("Valor en la caché:", ValueInCache)

			elapsed := time.Since(start)
			fmt.Printf("\nTiempo de búsqueda en la caché: %s \n", elapsed)
		}

		fmt.Println("\nPorcentaje de aciertos: ", (aciertos*100)/total, "%")

	}

}
