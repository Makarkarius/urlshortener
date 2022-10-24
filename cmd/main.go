package main

import (
	"log"
	"sync"
	"urlshortener/internal/server"
	"urlshortener/internal/urlstorage"
)

func main() {
	dbServer, err := server.NewServer(server.Config{
		Host: "localhost",
		Port: 8080,
		StorageCfg: urlstorage.Config{
			StorageType:       urlstorage.RuntimeStorage,
			UrlSize:           6,
			MaxKeyGenAttempts: 1e4,
			ShortenTimeoutSec: 10,
			ExpandTimeoutSec:  10,
			DBCfg: &urlstorage.DBStorageConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "",
				Password: "",
				Database: "",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := dbServer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	wg.Wait()
}
