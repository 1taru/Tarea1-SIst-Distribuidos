package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	addr      = flag.String("addr", "localhost:6379", "the address of the Redis server")
	message   = ""
	aciertos  = 0
	total     = 0
	serverURL = flag.String("server_url", "http://localhost:50051", "Server URL")
)

func main() {
	flag.Parse()

	redisClient := redis.NewClient(&redis.Options{
		Addr: *addr,
	})
	defer redisClient.Close()

	_, err := redisClient.Ping(context.Background()).Result()
	log.Printf("Conectándose al servidor Redis en %s", *addr)

	if err != nil {
		log.Fatal(err)
	}
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

		valueInCache, err := HandleGetFromCache(redisClient, message)
		if err != nil {
			log.Println(err)
		}

		if valueInCache == "" {
			start := time.Now()

			log.Println("No existe en la caché")

			// Obtener valor del servidor y guardarlo en caché si es necesario
			valueFromServer, err := GetValueFromServer(message)
			if err != nil {
				log.Println(err)
			}

			if valueFromServer != "" {
				// Guardar el valor en la caché
				err := redisClient.Set(context.Background(), message, valueFromServer, 0).Err()
				if err != nil {
					log.Println("Error al guardar en caché:", err)
				} else {
					log.Println("Valor almacenado en la caché")
				}
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

func HandleGetFromCache(redisClient *redis.Client, key string) (string, error) {
	valueInCache, err := redisClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil // Si la clave no existe en la caché, retornamos un valor vacío
	}
	if err != nil {
		return "", fmt.Errorf("error al obtener valor de la caché: %v", err)
	}
	return valueInCache, nil
}

func GetValueFromServer(key string) (string, error) {
	url := fmt.Sprintf("%s/getValue?key=%s", *serverURL, key)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to perform GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned non-OK status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var data struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON response: %v", err)
	}

	return data.Value, nil
}
