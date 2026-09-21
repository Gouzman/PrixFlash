// Appel au moteur de génération Rust (pubprix-engine) — API-05.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

var engineClient = &http.Client{Timeout: 30 * time.Second}

func engineBaseURL() string {
	if url := os.Getenv("ENGINE_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8081"
}

// GeneratedPost correspond à l'entité generated_post du modèle de données.
type GeneratedPost struct {
	ID        string    `json:"id"`
	DraftID   string    `json:"draft_id"`
	ImageURL  string    `json:"image_url,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	posts   = map[string]*GeneratedPost{}
	postsMu sync.RWMutex
)

type generateRequestPayload struct {
	Format string `json:"format"`
}

type engineGenerateRequest struct {
	PhotoURLs []string `json:"photo_urls"`
	Price     float64  `json:"price"`
	Name      string   `json:"name,omitempty"`
	Format    string   `json:"format"`
}

// engineGenerateResponse reflète la réponse de pubprix-engine : le moteur
// ne renvoie pas d'URL exploitable directement, mais le JPEG encodé en
// base64 (choix retenu côté moteur pour éviter d'y dupliquer des
// identifiants R2). C'est donc l'API qui uploade ce JPEG vers le même
// bucket que les photos sources et construit l'URL publique finale.
type engineGenerateResponse struct {
	Status      string  `json:"status"`
	ImageBase64 *string `json:"image_base64"`
	ContentType *string `json:"content_type"`
	Error       *string `json:"error"`
}

// POST /drafts/{id}/generate — API-05
func generateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	draft, ok := getDraft(id)
	if !ok {
		http.Error(w, "draft not found", http.StatusNotFound)
		return
	}
	if len(draft.Photos) < minPhotos {
		http.Error(w, "le brouillon n'a pas assez de photos", http.StatusUnprocessableEntity)
		return
	}
	if draft.Price <= 0 {
		http.Error(w, "le prix du brouillon est manquant", http.StatusUnprocessableEntity)
		return
	}

	format := "whatsapp"
	if r.Body != nil && r.ContentLength != 0 {
		var payload generateRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil && payload.Format != "" {
			format = payload.Format
		}
	}

	engineReq := engineGenerateRequest{
		PhotoURLs: draft.Photos,
		Price:     draft.Price,
		Name:      draft.Name,
		Format:    format,
	}
	body, err := json.Marshal(engineReq)
	if err != nil {
		http.Error(w, "requête invalide", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, engineBaseURL()+"/generate", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "impossible de préparer l'appel au moteur", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := engineClient.Do(req)
	if err != nil {
		http.Error(w, "moteur de génération indisponible", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "échec de la génération", http.StatusBadGateway)
		return
	}

	var engineResp engineGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&engineResp); err != nil {
		http.Error(w, "réponse du moteur invalide", http.StatusBadGateway)
		return
	}
	if engineResp.Status != "ok" || engineResp.ImageBase64 == nil {
		http.Error(w, "le moteur n'a pas pu générer le visuel", http.StatusBadGateway)
		return
	}

	imageURL, err := storeGeneratedImage(r.Context(), id, *engineResp.ImageBase64, engineResp.ContentType)
	if err != nil {
		if errors.Is(err, ErrStorageNotConfigured) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "échec du stockage du visuel généré", http.StatusBadGateway)
		}
		return
	}

	post := &GeneratedPost{
		ID:        uuid.NewString(),
		DraftID:   id,
		ImageURL:  imageURL,
		Status:    "ready",
		CreatedAt: time.Now().UTC(),
	}

	postsMu.Lock()
	posts[post.ID] = post
	postsMu.Unlock()

	writeJSON(w, http.StatusCreated, post)
}

// storeGeneratedImage décode le JPEG base64 renvoyé par le moteur et
// l'uploade vers le même stockage R2 que les photos sources, pour obtenir
// une URL publique exploitable par l'app (Image.network) et par les liens
// de partage (posts.go).
func storeGeneratedImage(ctx context.Context, draftID, imageBase64 string, contentType *string) (string, error) {
	if photoStorage == nil {
		return "", ErrStorageNotConfigured
	}

	decoded, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", fmt.Errorf("décodage du visuel généré: %w", err)
	}

	ct := "image/jpeg"
	if contentType != nil && *contentType != "" {
		ct = *contentType
	}

	key := fmt.Sprintf("posts/%s/%s.jpg", draftID, uuid.NewString())
	return photoStorage.upload(ctx, key, ct, bytes.NewReader(decoded))
}

func getPost(id string) (*GeneratedPost, bool) {
	postsMu.RLock()
	defer postsMu.RUnlock()
	post, ok := posts[id]
	return post, ok
}
