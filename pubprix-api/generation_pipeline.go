// Pipeline complet de génération pour le flux WhatsApp — IMG-03 :
// détourage (pubprix-detourage) -> fond de style + composition + overlay
// (pubprix-engine) -> stockage public (R2) -> envoi de l'image au
// vendeur (Twilio).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const whatsappGenerationTimeout = 90 * time.Second

// generationPipelineRunner est un point d'injection pour les tests :
// handleAttenteStyle appelle cette variable plutôt que
// startGenerationPipeline directement, ce qui permet de vérifier la
// transition d'état et le message renvoyé sans déclencher de vrais
// appels réseau (detourage/moteur/Twilio/R2 indisponibles en test
// unitaire). Le pipeline réel est validé par test d'intégration bout en
// bout (voir notes de commit) plutôt que par des doubles en Go.
var generationPipelineRunner = startGenerationPipeline

// styleKeyFor traduit le libellé affiché au vendeur (styleOptions,
// conversation.go) vers la valeur attendue par le moteur Rust.
func styleKeyFor(label string) string {
	switch label {
	case "Fond blanc":
		return "white"
	case "Fond couleur":
		return "color"
	case "Mise en scène":
		return "scene"
	default:
		return "white"
	}
}

// startGenerationPipeline lance le pipeline en arrière-plan et répond au
// vendeur (texte d'erreur ou image finale) une fois terminé, sans
// bloquer la réponse au webhook Twilio. sellersMu n'est jamais retenu
// pendant les appels réseau : on le reprend brièvement pour mettre à
// jour l'état une fois le résultat connu.
func startGenerationPipeline(seller *Seller, photoRefs []string, styleLabel string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), whatsappGenerationTimeout)
		defer cancel()

		imageURL, err := runGenerationPipeline(ctx, photoRefs, styleKeyFor(styleLabel), sanitizeForKey(seller.PhoneNumber))

		sellersMu.Lock()
		if err != nil {
			seller.State = StateReceptionPhotos
			// Le brouillon reste tel quel : le vendeur peut renvoyer une
			// meilleure photo sans tout recommencer.
		} else {
			seller.State = StatePret
			seller.Draft = nil
		}
		sellersMu.Unlock()

		if err != nil {
			log.Printf("pipeline génération (phone=%s): %v", seller.PhoneNumber, err)
			sendTwilioTextSafe(seller.PhoneNumber, userFacingGenerationError())
			return
		}

		log.Printf("pipeline génération (phone=%s): terminé -> %s", seller.PhoneNumber, imageURL)
		sendTwilioImageSafe(seller.PhoneNumber, imageURL)
	}()
}

// runGenerationPipeline fait le travail réel, séquentiellement : pas de
// verrou à tenir ici, seller n'est jamais touché directement (juste les
// références de photos, le style et un identifiant de vendeur assaini
// pour le namespacing des clés R2, copiés par l'appelant).
func runGenerationPipeline(ctx context.Context, photoRefs []string, engineStyle, ownerKey string) (string, error) {
	if photoStorage == nil {
		return "", ErrStorageNotConfigured
	}
	if twilioClient == nil {
		return "", ErrTwilioNotConfigured
	}

	detouredURLs := make([]string, 0, len(photoRefs))
	for i, ref := range photoRefs {
		original, contentType, err := twilioClient.DownloadMedia(ctx, ref)
		if err != nil {
			return "", fmt.Errorf("téléchargement photo %d: %w", i+1, err)
		}

		cutout, err := removeBackground(ctx, original, contentType)
		if err != nil {
			return "", fmt.Errorf("détourage photo %d: %w", i+1, err)
		}

		key := fmt.Sprintf("whatsapp-tmp/%s.png", uuid.NewString())
		if _, err := photoStorage.upload(ctx, key, "image/png", bytes.NewReader(cutout)); err != nil {
			return "", fmt.Errorf("stockage temporaire photo %d: %w", i+1, err)
		}
		signedURL, err := photoStorage.presignedGetURL(ctx, key, enginePhotoURLTTL)
		if err != nil {
			return "", fmt.Errorf("signature photo %d: %w", i+1, err)
		}
		detouredURLs = append(detouredURLs, signedURL)
	}

	engineReq := engineGenerateRequest{
		PhotoURLs: detouredURLs,
		Style:     engineStyle,
		Format:    "whatsapp",
	}
	body, err := json.Marshal(engineReq)
	if err != nil {
		return "", fmt.Errorf("préparation requête moteur: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, engineBaseURL()+"/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("préparation appel moteur: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := engineClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("appel moteur: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("moteur a répondu %d: %s", resp.StatusCode, string(respBody))
	}

	var engineResp engineGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&engineResp); err != nil {
		return "", fmt.Errorf("décodage réponse moteur: %w", err)
	}
	if engineResp.Status != "ok" || engineResp.ImageBase64 == nil {
		errMsg := ""
		if engineResp.Error != nil {
			errMsg = *engineResp.Error
		}
		return "", fmt.Errorf("le moteur n'a pas pu générer le visuel (status=%q error=%q)", engineResp.Status, errMsg)
	}

	finalURL, err := storeGeneratedImage(ctx, ownerKey, *engineResp.ImageBase64, engineResp.ContentType)
	if err != nil {
		return "", fmt.Errorf("stockage du visuel généré: %w", err)
	}
	return finalURL, nil
}

// userFacingGenerationError renvoie un message compréhensible plutôt que
// le détail technique de l'erreur — peu importe l'étape qui a échoué
// (détourage, moteur, stockage), le vendeur reçoit la même suggestion
// actionnable : reprendre une meilleure photo.
func userFacingGenerationError() string {
	return "La photo est trop sombre ou difficile à traiter, peux-tu la reprendre près d'une fenêtre et me la renvoyer ?"
}

func sendTwilioTextSafe(to, body string) {
	if twilioClient == nil {
		log.Printf("whatsapp reply (to=%s): %v", to, ErrTwilioNotConfigured)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := twilioClient.SendWhatsAppMessage(ctx, to, body); err != nil {
		log.Printf("whatsapp reply (to=%s): échec envoi texte: %v", to, err)
	}
}

func sendTwilioImageSafe(to, imageURL string) {
	if twilioClient == nil {
		log.Printf("whatsapp reply (to=%s): %v", to, ErrTwilioNotConfigured)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := twilioClient.SendWhatsAppImage(ctx, to, imageURL, "Voici ta publication ! 🎉"); err != nil {
		log.Printf("whatsapp reply (to=%s): échec envoi image: %v", to, err)
	}
}

// sanitizeForKey rend un numéro de téléphone utilisable comme segment de
// clé R2 (préfixe "whatsapp:" et "+" retirés) pour namespacer les
// visuels générés par vendeur.
func sanitizeForKey(s string) string {
	replacer := strings.NewReplacer("whatsapp:", "", "+", "", " ", "")
	return replacer.Replace(s)
}
