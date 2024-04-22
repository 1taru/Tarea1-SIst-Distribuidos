package main

import (
	"context"
	"encoding/csv"
	"flag"
	"io"

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
	//cache centralizado:

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	//cache con particionamiento
	/*
		redisClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{
				"localhost:7000",
				"localhost:7001",
				"localhost:7002",
			},
		})
	*/
	//cache con replicacion
	/*
		redisClient := redis.NewClient(&redis.Options{
			Addr: "localhost:7000",
		})
	*/
	defer redisClient.Close()

	// Verificación de conexión con Redis

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar al servidor Redis: %v", err)
	}

	//ping con con particionamiento
	/*
		err := redisClient.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
			return shard.Ping(ctx).Err()
		})

		if err != nil {
			panic(err)
		}
	*/
	// Obtención del valor desde Redis
	value, err := redisClient.Get(ctx, req.Key).Result()
	if err != nil {
		// Maneja el caso en el que la clave no existe en la caché
		if err == redis.Nil {
			log.Printf("clave '%s' no encontrada en la caché", req.Key)
			return nil, nil // Retorna nil para indicar que no se encontró la clave
		}
		return nil, fmt.Errorf("error al obtener el valor de la caché: %v", err)
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
	//cache centralizado:

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	//cache con replicacion
	/*
		redisClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{
				"localhost:7000",
				"localhost:7001",
				"localhost:7002",
			},
		})
	*/
	//cache con particionamiento
	/*
		redisClient := redis.NewClient(&redis.Options{
			Addr: "localhost:7000",
		})
	*/
	defer redisClient.Close()

	// Verifica la conexión con el servidor Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar al servidor Redis: %v", err)
	}

	// Ingresa el valor en Redis
	if err := redisClient.Append(ctx, req.Key, req.Value).Err(); err != nil {
		log.Fatalf("error al ingresar el valor en Redis: %v", err)
	}
	if err := redisClient.Expire(ctx, req.Key, 30*time.Second).Err(); err != nil {
		log.Fatalf("error al establecer el tiempo de expiración para la clave en Redis: %v", err)
	}
	// Retorna la respuesta del servidor gRPC
	return response, nil
}
func main() {
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("no se pudo conectar al servidor gRPC: %v", err)
	}
	defer conn.Close()

	client := pb.NewCacheServiceClient(conn)
	cacheClient := &CacheClient{cc: client}
	ctx := context.Background()

	file, err := os.Open("/home/taru/Downloads/t1_distribuidos/Tarea1-SIst-Distribuidos//Universitario.csv")
	if err != nil {
		log.Fatalf("error al abrir el archivo CSV: %v", err)
	}
	defer file.Close()
	// Lee el archivo CSV como un lector CSV
	reader := csv.NewReader(file)
	reader.Comma = ',' // Especifica el delimitador como '|'
	var sumElapsed time.Duration
	for {
		record, err := reader.Read()
		if err != nil {
			// Verifica si se alcanzó el final del archivo
			if err == io.EOF {
				break
			}
			log.Fatalf("error al leer el archivo CSV: %v", err)
		}
		message := record[21]
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
				Value: "Ingresado a Cache",
			}

			// Llama a la función SetInCache para almacenar el valor en la caché
			_, err := cacheClient.SetInCache(ctx, setRequest)
			if err != nil {
				log.Fatalf("error al llamar a SetInCache: %v", err)
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
			sumElapsed += elapsed
			fmt.Printf("\nTiempo de búsqueda en la caché: %s \n", elapsed)
		}

		fmt.Println("\nPorcentaje de aciertos: ", (aciertos*100)/total, "%")

	}
	microseconds := sumElapsed.Microseconds()
	microsecondsInt := int(microseconds)

	fmt.Println("Tiempo promedio de búsqueda en la caché: ", microsecondsInt/total, "us")
	fmt.Println("----------------------------")
	fmt.Println("Fin del archivo CSV")

}
