// Upload des photos d'un brouillon vers le stockage objets — API-03.
package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
)

const (
	minPhotos    = 2
	maxPhotos    = 6
	maxUploadMem = 32 << 20 // 32 MiB gardés en mémoire par ParseMultipartForm
)

var allowedPhotoTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// photoStorage est initialisé au démarrage depuis les variables
// d'environnement (voir storage.go). Il peut rester nil si l'API tourne
// sans configuration S3 pour du développement local sur les autres
// endpoints ; l'upload de photos échoue alors proprement en 500.
var photoStorage *PhotoStorage

// POST /drafts/{id}/photos — API-03
func uploadPhotosHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, ok := getDraft(id); !ok {
		http.Error(w, "draft not found", http.StatusNotFound)
		return
	}

	if photoStorage == nil {
		http.Error(w, ErrStorageNotConfigured.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(maxUploadMem); err != nil {
		http.Error(w, "invalid multipart body", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	files := r.MultipartForm.File["photos"]
	if len(files) < minPhotos || len(files) > maxPhotos {
		http.Error(w, fmt.Sprintf("entre %d et %d photos sont attendues", minPhotos, maxPhotos), http.StatusBadRequest)
		return
	}

	urls := make([]string, 0, len(files))
	for _, fh := range files {
		contentType := fh.Header.Get("Content-Type")
		if !allowedPhotoTypes[contentType] {
			http.Error(w, fmt.Sprintf("type de fichier non supporté: %s", contentType), http.StatusBadRequest)
			return
		}

		file, err := fh.Open()
		if err != nil {
			http.Error(w, "impossible de lire une photo", http.StatusBadRequest)
			return
		}

		key := fmt.Sprintf("drafts/%s/%s%s", id, uuid.NewString(), filepath.Ext(fh.Filename))
		url, err := photoStorage.upload(r.Context(), key, contentType, file)
		_ = file.Close()
		if err != nil {
			http.Error(w, "échec de l'upload vers le stockage", http.StatusBadGateway)
			return
		}
		urls = append(urls, url)
	}

	draftsMu.Lock()
	if draft, ok := drafts[id]; ok {
		draft.Photos = urls
	}
	draftsMu.Unlock()

	writeJSON(w, http.StatusOK, map[string][]string{"photos": urls})
}
