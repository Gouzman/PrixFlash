// Pubprix API — orchestre l'app Flutter, le stockage des photos et le
// moteur de génération Rust. Voir le cahier des charges technique,
// section "API — endpoints principaux du MVP".
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func main() {
	storage, err := newPhotoStorageFromEnv(context.Background())
	if err != nil {
		if errors.Is(err, ErrStorageNotConfigured) {
			log.Printf("attention: %v — l'upload de photos sera indisponible", err)
		} else {
			log.Fatalf("initialisation du stockage S3: %v", err)
		}
	}
	photoStorage = storage

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("POST /drafts", createDraftHandler)
	mux.HandleFunc("PATCH /drafts/{id}", updateDraftHandler)
	mux.HandleFunc("POST /drafts/{id}/photos", uploadPhotosHandler)
	mux.HandleFunc("POST /drafts/{id}/generate", generateHandler)
	mux.HandleFunc("GET /posts/{id}/share-links", shareLinksHandler)

	addr := ":8080"
	log.Printf("Pubprix API à l'écoute sur %s", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}
