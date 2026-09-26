// Webhook entrant WhatsApp (Twilio) — WA-02.
//
// Flux "aller-retour" minimal pour prouver que ça marche de bout en bout
// avec le bac à sable : on reçoit le message, on le logue, et on répond
// par un echo via l'API Twilio sortante. Pas de machine à états de
// conversation à ce stade (voir WA-03).
package main

import (
	"log"
	"net/http"
	"strconv"
)

const echoReplyText = "Bonjour ! Pubprix devient VitrinePro"

// twilioClient est initialisé au démarrage depuis les variables
// d'environnement (voir twilio.go). Peut rester nil pour du développement
// local sans Twilio configuré ; le webhook logue alors le message reçu
// sans pouvoir y répondre.
var twilioClient *TwilioClient

// POST /webhooks/twilio/whatsapp
//
// Twilio envoie les messages entrants en x-www-form-urlencoded, avec
// notamment les champs From (expéditeur, "whatsapp:+225..."), Body (texte)
// et NumMedia (nombre de pièces jointes).
func twilioWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form body", http.StatusBadRequest)
		return
	}

	from := r.FormValue("From")
	body := r.FormValue("Body")
	numMedia, _ := strconv.Atoi(r.FormValue("NumMedia"))

	log.Printf("whatsapp in: from=%s hasMedia=%t numMedia=%d body=%q", from, numMedia > 0, numMedia, body)

	switch {
	case from == "":
		log.Printf("whatsapp reply: champ From absent, impossible de répondre")
	case twilioClient == nil:
		log.Printf("whatsapp reply (to=%s): %v", from, ErrTwilioNotConfigured)
	default:
		if err := twilioClient.SendWhatsAppMessage(r.Context(), from, echoReplyText); err != nil {
			log.Printf("whatsapp reply (to=%s): échec: %v", from, err)
		} else {
			log.Printf("whatsapp reply (to=%s): envoyé", from)
		}
	}

	// TwiML vide : la réponse réelle part via l'appel API Twilio
	// ci-dessus, pas via une action TwiML synchrone.
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Response></Response>`))
}
