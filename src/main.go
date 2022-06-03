package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
)

var (
	urlMap      map[string]string
	urlMapMutex sync.Mutex
	keyMap      map[string]string
	keyMapMutex sync.Mutex
)

type shortenRequest struct {
	Url string
}

type shortenResponse struct {
	Url string
	Key string
}

func generateKey(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var request shortenRequest
		if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}

		keyMapMutex.Lock()
		key, found := keyMap[request.Url]
		keyMapMutex.Unlock()

		if !found {
			var e error
			key, e = generateKey(3)
			if e != nil {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}
			urlMapMutex.Lock()
			urlMap[key] = request.Url
			urlMapMutex.Unlock()

			keyMapMutex.Lock()
			keyMap[request.Url] = key
			keyMapMutex.Unlock()
		}

		response := shortenResponse{Url: request.Url, Key: key}

		w.Header().Set("Content-Type", "application/json")
		if e := json.NewEncoder(w).Encode(response); e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodGet:
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		vars := mux.Vars(r)
		fmt.Println(vars["key"])

		urlMapMutex.Lock()
		newUrl, found := urlMap[vars["key"]]
		urlMapMutex.Unlock()

		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Location", newUrl)
		http.Redirect(w, r, newUrl, http.StatusFound)
	case http.MethodPost:
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func main() {
	port := flag.Int("port", 8080, "port number")
	flag.Parse()

	router := mux.NewRouter()
	urlMap = make(map[string]string)
	keyMap = make(map[string]string)

	router.HandleFunc("/shorten", shortenHandler)
	router.HandleFunc("/go/{key}", goHandler)
	fmt.Println(fmt.Sprintf("localhost:%d", *port))
	e := http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), router)
	if e != nil {
		log.Fatal(e)
	}
}
