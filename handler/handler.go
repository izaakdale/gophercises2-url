package urlshort

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/go-yaml/yaml"
)

func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if destUrl, found := pathsToUrls[r.URL.Path]; found {
			http.Redirect(w, r, destUrl, http.StatusFound)
		}
		fallback.ServeHTTP(w, r)
	}
}

type UrlPath struct {
	Path string
	Url  string
}

func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {

	var yamlData []UrlPath
	if err := yaml.Unmarshal(yml, &yamlData); err != nil {
		return nil, err
	}
	yamlMap := make(map[string]string)
	for _, value := range yamlData {
		yamlMap[value.Path] = value.Url
	}

	return MapHandler(yamlMap, fallback), nil
}

func JSONHandler(jsonBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var jsonData []UrlPath
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return nil, err
	}
	jsonMap := make(map[string]string)
	for _, value := range jsonData {
		jsonMap[value.Path] = value.Url
	}
	return func(w http.ResponseWriter, r *http.Request) {
		destUrl, found := jsonMap[r.URL.Path]
		if found {
			http.Redirect(w, r, destUrl, http.StatusFound)
		}
		fallback.ServeHTTP(w, r)
	}, nil
}

func RedisHandler(client *redis.Client, ctx *context.Context, fallback http.Handler) (http.HandlerFunc, error) {

	return func(w http.ResponseWriter, r *http.Request) {
		url, err := client.Get(*ctx, r.URL.Path).Result()
		if err == redis.Nil {
			// redis nil means key does not exist, use fallback
			fallback.ServeHTTP(w, r)
			return
		}
		// use url found in path
		http.Redirect(w, r, url, http.StatusFound)

	}, nil

}
