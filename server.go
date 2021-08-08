package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func GetInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Starting /get_info")
	if r.Method == http.MethodPost {
		urls, err := parseBody(w, r)
		if err != nil {
			return
		}
		if len(urls) > 20 {
			log.Println("Too many urls in request")
			w.WriteHeader(400)
			return
		}
		respMap := make(map[string]string)
		respMap, err = makeRequests(urls)
		if err != nil {
			w.WriteHeader(500)
			_, err = w.Write([]byte(fmt.Sprintf("Error: %s", err)))
			OnErr(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		obj, err := json.Marshal(respMap)
		if err != nil {
			log.Printf("Marshaling error, %v", err)
			w.WriteHeader(500)
			return
		}
		_, err = w.Write(obj)
		if err != nil {
			log.Printf("Error while writing in ResponseWriter: %v", err)
			w.WriteHeader(500)
			return
		}
	} else {
		fmt.Println("Method not allowed")
		w.WriteHeader(405)
	}
}

func Serve(ctx context.Context) (err error) {

	mux := http.NewServeMux()
	mux.Handle("/", maxClientsMiddleware(http.HandlerFunc(GetInfo), 100))

	srv := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen:%s\n", err)
		}
	}()

	log.Printf("Server started")

	<-ctx.Done()

	log.Printf("Server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server shutdown failed:%s", err)
	}

	log.Printf("Server shutting down")

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}
