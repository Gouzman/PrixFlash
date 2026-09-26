// Client Twilio — envoi de messages WhatsApp sortants (bac à sable en
// dev, numéro WhatsApp Business vérifié en prod) — WA-02.
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

// TwilioClient encapsule l'appel à l'API REST Twilio (ressource Messages)
// pour l'envoi de messages WhatsApp. Authentification par Account SID +
// Auth Token classique (Basic Auth) — c'est la seule paire qui a
// effectivement fonctionné en test ; l'API Key associée au compte n'avait
// pas son secret correctement renseigné.
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
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// SendWhatsAppMessage envoie un message WhatsApp sortant via l'API REST
// Twilio. `to` est déjà au format "whatsapp:+225XXXXXXXXX" (tel que reçu
// dans le champ From du webhook entrant).
func (t *TwilioClient) SendWhatsAppMessage(ctx context.Context, to, body string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	form := url.Values{}
	form.Set("From", t.fromNumber)
	form.Set("To", to)
	form.Set("Body", body)

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
