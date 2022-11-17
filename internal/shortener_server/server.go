package shortener_server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"time"
	"urlshortener/internal/urlstorage"
)

type Config struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ReadTimeoutSec  int    `json:"readTimeoutSec"`
	WriteTimeoutSec int    `json:"writeTimeoutSec"`

	LoggerCfg  zap.Config        `json:"loggerCfg"`
	StorageCfg urlstorage.Config `json:"storageCfg"`
}

type Server struct {
	cfg     Config
	server  http.Server
	storage urlstorage.Storage
	logger  *zap.Logger
}

func NewServer(cfg Config) (*Server, error) {
	storage, err := urlstorage.NewStorage(cfg.StorageCfg)
	if err != nil {
		return nil, err
	}
	logger, err := cfg.LoggerCfg.Build()
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg: cfg,
		server: http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      nil,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		storage: storage,
		logger:  logger,
	}, nil
}

func (s *Server) Run() error {
	router := mux.NewRouter()
	router.HandleFunc("/shorten", s.handle(handleShorten))
	router.HandleFunc("/go/{key}", s.handle(handleGo))
	s.server.Handler = router

	defer func() {
		s.logger.Info("server stopped")
		_ = s.logger.Sync()
	}()
	s.logger.Info("starting server")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

type shortenRequest struct {
	Url string
}

type shortenResponse struct {
	Url string
	Key string
}

func (s *Server) handle(f func(s *Server, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(s, w, r)
	}
}

func handleShorten(s *Server, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.logger.Info("handling shorten request")
		var request shortenRequest
		if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}

		key, err := s.storage.Shorten(r.Context(), request.Url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := shortenResponse{Url: request.Url, Key: key}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodGet:
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func handleGo(s *Server, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.logger.Info("handling go request")
		vars := mux.Vars(r)
		url, err := s.storage.Expand(r.Context(), vars["key"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if url == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Location", url)
		http.Redirect(w, r, url, http.StatusFound)
	case http.MethodPost:
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}
