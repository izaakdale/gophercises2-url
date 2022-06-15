package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-redis/redis/v9"
	urlshort "github.com/izaakdale/urlShortner/handler"
)

func main() {
	mux := defaultMux()
	redisClient, ctx := redisClient()

	yamlFileName := flag.String("yaml", "default.yaml", "Flag defines a yaml file to use as url shortcut config")
	jsonFileName := flag.String("json", "default.json", "Flag defines a json file to use as url shortcut config")

	flag.Parse()

	yaml, err := ioutil.ReadFile(*yamlFileName)
	if err != nil {
		fmt.Printf("Failed to read yaml file %s", err.Error())
		os.Exit(1)
	}

	json, err := ioutil.ReadFile(*jsonFileName)
	if err != nil {
		fmt.Printf("Failed to read yaml file %s", err.Error())
		os.Exit(1)
	}

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	yamlHandler, err := urlshort.YAMLHandler(yaml, mapHandler)
	if err != nil {
		panic(err)
	}
	jsonHandler, err := urlshort.JSONHandler(json, yamlHandler)
	if err != nil {
		panic(err)
	}
	redisHandler, err := urlshort.RedisHandler(redisClient, ctx, jsonHandler)
	if err != nil {
		panic(err)
	}
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", redisHandler)
}

func redisClient() (*redis.Client, *context.Context) {
	ctx := context.Background()
	return redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}), &ctx
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
