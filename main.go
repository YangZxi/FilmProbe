package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

func test() {
	filmInfos, _ := _4ksjHandler("流浪地球", target[2])
	log.Printf("filmList: %v", filmInfos)
	log.Fatalf("")
}

func main() {
	//test()
	port := "8080"
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		word := query.Get("w")
		log.Print("w: ", word)
		if word == "" {
			http.Error(w, "Missing 'word' parameter", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		filmList := Search(word)
		if err := json.NewEncoder(w).Encode(filmList); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		targetURL := query.Get("url")
		if targetURL == "" {
			http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
			return
		}
		host := r.Host
		referer := r.Header.Get("Referer")
		index := strings.Index(referer, host)
		if index == -1 || index > 8 {
			http.Error(w, "Request Forbidden", http.StatusForbidden)
			return
		}

		headers := map[string]string{}
		body, resp := sendRequest("GET", targetURL, headers, nil)
		if body == nil {
			http.Error(w, "proxy failed", http.StatusInternalServerError)
			return
		}
		//println(response.Body)

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, strings.NewReader(string(body)))
	})

	fmt.Printf("Starting server at http://127.0.0.1:%s \n", port)
	if err := http.ListenAndServe(":"+port, recoverMiddleware(mux)); err != nil {
		fmt.Println("Failed to start server:", err)
	}

}
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
