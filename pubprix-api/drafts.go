// Brouillons de produit — API-02 et API-04.
// Voir le cahier des charges technique, section "API — endpoints
// principaux du MVP".
package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Draft correspond à l'entité product_draft du modèle de données.
type Draft struct {
	ID           string    `json:"id"`
	SessionToken string    `json:"session_token"`
	Name         string    `json:"name,omitempty"`
	Price        float64   `json:"price,omitempty"`
	Currency     string    `json:"currency,omitempty"`
	Photos       []string  `json:"photos,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Stockage en mémoire pour le MVP — à remplacer par Postgres ou équivalent
// dès que la persistance au-delà d'une session est nécessaire (voir
// backlog, ticket API-07 pour le nettoyage automatique).
var (
	drafts   = map[string]*Draft{}
	draftsMu sync.RWMutex
)

// POST /drafts — API-02
func createDraftHandler(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		sessionToken = uuid.NewString()
	}

	draft := &Draft{
		ID:           uuid.NewString(),
		SessionToken: sessionToken,
		Currency:     "FCFA",
		CreatedAt:    time.Now().UTC(),
	}

	draftsMu.Lock()
	drafts[draft.ID] = draft
	draftsMu.Unlock()

	w.Header().Set("X-Session-Token", sessionToken)
	writeJSON(w, http.StatusCreated, draft)
}

// PATCH /drafts/{id} — API-04 (prix, nom du produit)
type updateDraftPayload struct {
	Name  *string  `json:"name"`
	Price *float64 `json:"price"`
}

func updateDraftHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	draftsMu.Lock()
	defer draftsMu.Unlock()

	draft, ok := drafts[id]
	if !ok {
		http.Error(w, "draft not found", http.StatusNotFound)
		return
	}

	var payload updateDraftPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if payload.Name != nil {
		draft.Name = *payload.Name
	}
	if payload.Price != nil {
		if *payload.Price <= 0 {
			http.Error(w, "price must be positive", http.StatusBadRequest)
			return
		}
		draft.Price = *payload.Price
	}

	writeJSON(w, http.StatusOK, draft)
}

func getDraft(id string) (*Draft, bool) {
	draftsMu.RLock()
	defer draftsMu.RUnlock()
	draft, ok := drafts[id]
	return draft, ok
}
