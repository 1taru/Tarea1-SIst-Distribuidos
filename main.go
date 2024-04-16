package main

import (
	"fmt"
	"net/http"
    "github.com/go-redis/redis"
	"rsc.io/quote"
	
)

func handler(w http.ResponseWriter, r *http.Request) {
	
	fmt.Fprintf(w, "¡Hola, mundo!")
}
func handler2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "chao, mundo!")
}
func get_name(w redis.Client) string {
	val, err := w.Get("name").Result()
	if err != nil {
		panic(err)
	}
	return val
}
func main() {
	fmt.Println(quote.Go())
	client := redis.NewClient(&redis.Options{
        Addr:     "redis-17721.c56.east-us.azure.cloud.redislabs.com:17721",
        Password: "5u0q8oOEYwgqL8WSHRlA9yEa6d4hmt6n", // no password set
        DB:       0,  // use default DB
    })

    // Ping Redis to check if the connection is working
    _, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }
	
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", handler)
	http.HandleFunc("/chao", handler2)
	fmt.Println("Servidor escuchando en http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}