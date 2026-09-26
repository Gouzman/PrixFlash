// Webhook entrant WhatsApp (Twilio) — WA-02/WA-03.
//
// Reçoit le message, en extrait ce qu'il faut pour la machine à états de
// conversation (conversation.go), puis répond via l'API Twilio sortante
// (pas de TwiML synchrone).
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// twilioClient est initialisé au démarrage depuis les variables
// d'environnement (voir twilio.go). Peut rester nil pour du développement
// local sans Twilio configuré ; le webhook logue alors la réponse calculée
// sans pouvoir l'envoyer.
var twilioClient *TwilioClient

// POST /webhooks/twilio/whatsapp
//
// Twilio envoie les messages entrants en x-www-form-urlencoded, avec
// notamment les champs From (expéditeur, "whatsapp:+225..."), Body (texte),
// NumMedia (nombre de pièces jointes) et MediaUrl0..N-1 (leurs URLs).
func twilioWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form body", http.StatusBadRequest)
		return
	}

	from := r.FormValue("From")
	body := r.FormValue("Body")
	numMedia, _ := strconv.Atoi(r.FormValue("NumMedia"))

	mediaRefs := make([]string, 0, numMedia)
	for i := 0; i < numMedia; i++ {
		if url := r.FormValue(fmt.Sprintf("MediaUrl%d", i)); url != "" {
			mediaRefs = append(mediaRefs, url)
		}
	}

	log.Printf("whatsapp in: from=%s hasMedia=%t numMedia=%d body=%q", from, len(mediaRefs) > 0, numMedia, body)

	if from == "" {
		log.Printf("whatsapp reply: champ From absent, impossible de répondre")
		writeEmptyTwiML(w)
		return
	}

	reply := HandleIncomingMessage(IncomingMessage{From: from, Body: body, MediaRefs: mediaRefs})

	switch {
	case twilioClient == nil:
		log.Printf("whatsapp reply (to=%s): %v (réponse calculée: %q)", from, ErrTwilioNotConfigured, reply)
	default:
		if err := twilioClient.SendWhatsAppMessage(r.Context(), from, reply); err != nil {
			log.Printf("whatsapp reply (to=%s): échec: %v (réponse calculée: %q)", from, err, reply)
		} else {
			log.Printf("whatsapp reply (to=%s): envoyé: %q", from, reply)
		}
	}

	writeEmptyTwiML(w)
}

func writeEmptyTwiML(w http.ResponseWriter) {
	// TwiML vide : la réponse réelle part via l'appel API Twilio
	// ci-dessus, pas via une action TwiML synchrone.
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Response></Response>`))
}
