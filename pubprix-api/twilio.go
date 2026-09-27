// Client Twilio — envoi/réception de médias WhatsApp (bac à sable en
// dev, numéro WhatsApp Business vérifié en prod) — WA-02, IMG-03.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ErrTwilioNotConfigured signale que les variables d'environnement Twilio
// n'ont pas été fournies.
var ErrTwilioNotConfigured = errors.New("Twilio non configuré : TWILIO_ACCOUNT_SID et TWILIO_AUTH_TOKEN sont requis")

// defaultSandboxFromNumber est le numéro WhatsApp du bac à sable Twilio,
// identique pour tous les comptes en mode sandbox.
const defaultSandboxFromNumber = "whatsapp:+14155238886"

// TwilioClient encapsule l'accès à l'API REST Twilio : envoi de messages
// WhatsApp (texte ou image) et téléchargement des pièces jointes reçues.
// Authentification par Account SID + Auth Token classique (Basic Auth) —
// c'est la seule paire qui a effectivement fonctionné en test ; l'API Key
// associée au compte n'avait pas son secret correctement renseigné.
type TwilioClient struct {
	accountSID string
	authToken  string
	fromNumber string // format "whatsapp:+14155238886"
	httpClient *http.Client
}

func newTwilioClientFromEnv() (*TwilioClient, error) {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	if accountSID == "" || authToken == "" {
		return nil, ErrTwilioNotConfigured
	}

	fromNumber := os.Getenv("TWILIO_WHATSAPP_FROM")
	if fromNumber == "" {
		fromNumber = defaultSandboxFromNumber
	} else if !strings.HasPrefix(fromNumber, "whatsapp:") {
		fromNumber = "whatsapp:" + fromNumber
	}

	return &TwilioClient{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// SendWhatsAppMessage envoie un message WhatsApp texte. `to` est déjà au
// format "whatsapp:+225XXXXXXXXX" (tel que reçu dans le champ From du
// webhook entrant).
func (t *TwilioClient) SendWhatsAppMessage(ctx context.Context, to, body string) error {
	form := url.Values{}
	form.Set("Body", body)
	return t.sendMessage(ctx, to, form)
}

// SendWhatsAppImage envoie un message WhatsApp avec une image en pièce
// jointe (+ légende optionnelle). `mediaURL` doit être une URL
// publiquement accessible : Twilio la télécharge lui-même côté serveur
// pour la transmettre via WhatsApp (paramètre `MediaUrl` de l'API
// Messages, documenté pour l'envoi de media WhatsApp sortant).
func (t *TwilioClient) SendWhatsAppImage(ctx context.Context, to, mediaURL, caption string) error {
	form := url.Values{}
	if caption != "" {
		form.Set("Body", caption)
	}
	form.Set("MediaUrl", mediaURL)
	return t.sendMessage(ctx, to, form)
}

func (t *TwilioClient) sendMessage(ctx context.Context, to string, form url.Values) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	form.Set("From", t.fromNumber)
	form.Set("To", to)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.accountSID, t.authToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Twilio a répondu %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// DownloadMedia télécharge une pièce jointe reçue (ex. MediaUrl0 d'un
// webhook entrant). Ces URLs sont protégées : Twilio exige la même
// authentification Basic Auth (Account SID + Auth Token) que pour l'API
// REST elle-même. Renvoie les octets et le Content-Type annoncé par
// Twilio.
func (t *TwilioClient) DownloadMedia(ctx context.Context, mediaURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("User-Agent", "pubprix-api/1.0 (+https://github.com/pubprix)")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("téléchargement média Twilio: %d: %s", resp.StatusCode, string(body))
	}

	return body, resp.Header.Get("Content-Type"), nil
}
