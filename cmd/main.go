package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"urlshortener/internal/shortener_server"
)

// TODO: add load balancer
// TODO: add containerization

var cfgPath string

func init() {
	const (
		defaultCfgPath = ""
		usage          = "config path"
	)
	flag.StringVar(&cfgPath, "config", defaultCfgPath, usage)
	flag.StringVar(&cfgPath, "c", defaultCfgPath, usage+" (shorthand)")
}

func getCfg(path string) (result shortener_server.Config, _ error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return shortener_server.Config{}, err
	}
	err = json.Unmarshal(raw, &result)
	if err != nil {
		return shortener_server.Config{}, err
	}
	return result, nil
}

func main() {
	flag.Parse()
	if len(cfgPath) == 0 {
		log.Fatal("config path isn't provided")
	}
	cfg, err := getCfg(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	server, err := shortener_server.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
	<-idleConnsClosed
}
