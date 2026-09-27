// Client pour pubprix-detourage (microservice Python de détourage de
// fond) — IMG-01/IMG-03.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"
)

var detourageClient = &http.Client{Timeout: 60 * time.Second}

func detourageBaseURL() string {
	if url := os.Getenv("DETOURAGE_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8082"
}

// removeBackground envoie l'image à pubprix-detourage et renvoie le PNG à
// fond transparent résultant.
//
// Le champ Content-Type de la partie multipart est fixé explicitement
// (via CreatePart plutôt que CreateFormFile, qui met
// "application/octet-stream" par défaut) — même bug que celui déjà
// rencontré côté client Dart : sans ça, le service de détourage rejette
// l'upload en 400 "type de fichier non supporté".
func removeBackground(ctx context.Context, imageBytes []byte, contentType string) ([]byte, error) {
	if contentType == "" {
		contentType = "image/jpeg"
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="photo"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, fmt.Errorf("préparation requête detourage: %w", err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		return nil, fmt.Errorf("préparation requête detourage: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("préparation requête detourage: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, detourageBaseURL()+"/remove-background", &buf)
	if err != nil {
		return nil, fmt.Errorf("préparation requête detourage: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := detourageClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel detourage: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse detourage: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("detourage a répondu %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}
