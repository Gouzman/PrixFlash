// Liens de partage pré-remplis à partir d'une publication générée — API-06.
package main

import (
	"net/http"
	"net/url"
)

// GET /posts/{id}/share-links — API-06
func shareLinksHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	post, ok := getPost(id)
	if !ok {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	if post.ImageURL == "" {
		http.Error(w, "le visuel n'est pas encore disponible", http.StatusConflict)
		return
	}

	message := "Découvrez ma publication : " + post.ImageURL

	links := map[string]string{
		"whatsapp": "https://wa.me/?text=" + url.QueryEscape(message),
		"facebook": "https://www.facebook.com/sharer/sharer.php?u=" + url.QueryEscape(post.ImageURL),
		// Instagram n'expose pas d'API de partage web par URL : l'app doit
		// proposer l'image via le partage natif du téléphone (share sheet),
		// on renvoie donc directement l'URL du visuel à partager.
		"instagram": post.ImageURL,
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"post_id": post.ID,
		"links":   links,
	})
}
