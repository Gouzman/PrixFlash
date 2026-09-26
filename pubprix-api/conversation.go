// Machine à états de conversation WhatsApp par vendeur — WA-03.
//
// Stockage en mémoire pour l'instant (comme les drafts à l'époque de
// Pubprix) : une entrée par numéro de téléphone, à remplacer par une
// vraie base de données avec INF-06.
package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type SellerState string

const (
	StateAttenteNomBoutique SellerState = "ATTENTE_NOM_BOUTIQUE"
	StateAttenteLogo        SellerState = "ATTENTE_LOGO"
	StatePret               SellerState = "PRET"
	StateReceptionPhotos    SellerState = "RECEPTION_PHOTOS"
)

// maxDraftPhotos borne le nombre de photos par publication WhatsApp (1 à
// 6) — distinct de minPhotos/maxPhotos (photos.go), qui bornent l'upload
// multipart de l'ancien flux Flutter (2 à 6).
const maxDraftPhotos = 6

const readyMessage = "Envoie-moi 1 à 6 photos de ton produit, avec le prix si tu veux."

// SellerDraft est la publication en préparation pour un vendeur : juste
// les références des photos reçues et les notes texte qui les
// accompagnent (prix, détails...) — pas de traitement d'image à ce stade
// (RS/API à venir).
type SellerDraft struct {
	Photos []string
	Notes  []string
}

// Seller représente un vendeur WhatsApp et son avancement dans la
// conversation.
type Seller struct {
	PhoneNumber string
	State       SellerState
	ShopName    string
	LogoRef     string
	Draft       *SellerDraft
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	sellers   = map[string]*Seller{}
	sellersMu sync.Mutex
)

// IncomingMessage regroupe ce qu'on extrait du webhook Twilio,
// indépendamment du format de transport.
type IncomingMessage struct {
	From      string
	Body      string
	MediaRefs []string // URLs Twilio des pièces jointes, dans l'ordre reçu
}

// HandleIncomingMessage fait avancer la machine à états pour le numéro
// donné et renvoie le texte à répondre. Tout se fait sous verrou : un
// vendeur ne traite qu'un message à la fois, ce qui évite les courses
// entre deux messages rapprochés du même numéro.
func HandleIncomingMessage(msg IncomingMessage) string {
	sellersMu.Lock()
	defer sellersMu.Unlock()

	body := strings.TrimSpace(msg.Body)
	hasMedia := len(msg.MediaRefs) > 0
	now := time.Now().UTC()

	seller, known := sellers[msg.From]
	if !known {
		// État NOUVEAU : pas persisté en tant que tel, c'est simplement
		// l'absence d'entrée pour ce numéro. Le premier message ne sert
		// que de déclencheur — son contenu n'est pas interprété.
		seller = &Seller{
			PhoneNumber: msg.From,
			State:       StateAttenteNomBoutique,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		sellers[msg.From] = seller
		return "Bienvenue sur VitrinePro ! 👋 Comment s'appelle ta boutique ?"
	}
	seller.UpdatedAt = now

	switch seller.State {
	case StateAttenteNomBoutique:
		return handleAttenteNomBoutique(seller, body)
	case StateAttenteLogo:
		return handleAttenteLogo(seller, body, hasMedia, msg.MediaRefs)
	case StatePret:
		return handlePret(seller, body, hasMedia, msg.MediaRefs)
	case StateReceptionPhotos:
		return handleReceptionPhotos(seller, body, hasMedia, msg.MediaRefs)
	default:
		// Ne devrait pas arriver ; on ne perd pas le vendeur pour autant.
		seller.State = StateAttenteNomBoutique
		return "Reprenons : comment s'appelle ta boutique ?"
	}
}

func handleAttenteNomBoutique(seller *Seller, body string) string {
	if body == "" {
		return "Je n'ai pas compris. Quel est le nom de ta boutique ?"
	}
	seller.ShopName = body
	seller.State = StateAttenteLogo
	return fmt.Sprintf(
		"Boutique « %s » enregistrée ! Veux-tu ajouter un logo ? Envoie une photo, ou réponds \"non\" pour passer.",
		body,
	)
}

func handleAttenteLogo(seller *Seller, body string, hasMedia bool, mediaRefs []string) string {
	switch {
	case hasMedia:
		seller.LogoRef = mediaRefs[0]
		seller.State = StatePret
		return "Logo enregistré ! " + readyMessage
	case isSkip(body):
		seller.State = StatePret
		return readyMessage
	default:
		return "Envoie une photo pour ton logo, ou réponds \"non\" pour passer cette étape."
	}
}

func handlePret(seller *Seller, body string, hasMedia bool, mediaRefs []string) string {
	if !hasMedia {
		return readyMessage
	}
	// Un vendeur déjà enregistré qui envoie une photo passe directement
	// en réception de photos, sans repasser par l'accueil.
	seller.State = StateReceptionPhotos
	seller.Draft = &SellerDraft{}
	return addToDraft(seller.Draft, body, mediaRefs)
}

func handleReceptionPhotos(seller *Seller, body string, hasMedia bool, mediaRefs []string) string {
	if seller.Draft == nil {
		seller.Draft = &SellerDraft{}
	}
	if !hasMedia && body == "" {
		return "Envoie une photo ou un prix pour continuer ta publication."
	}
	return addToDraft(seller.Draft, body, mediaRefs)
}

// addToDraft ajoute les photos (jusqu'à maxDraftPhotos) et la note texte
// éventuelle au brouillon en cours, puis renvoie le message de statut
// approprié.
func addToDraft(draft *SellerDraft, body string, mediaRefs []string) string {
	added := 0
	for _, ref := range mediaRefs {
		if len(draft.Photos) >= maxDraftPhotos {
			break
		}
		draft.Photos = append(draft.Photos, ref)
		added++
	}
	if body != "" {
		draft.Notes = append(draft.Notes, body)
	}

	switch {
	case added == 0 && len(mediaRefs) > 0:
		return fmt.Sprintf(
			"Tu as déjà envoyé %d photos, c'est le maximum pour cette publication. Réponds avec le prix si besoin.",
			maxDraftPhotos,
		)
	case len(draft.Photos) > 0:
		return fmt.Sprintf(
			"%d/%d photo(s) reçue(s). Envoie-en d'autres ou indique le prix quand tu es prêt.",
			len(draft.Photos), maxDraftPhotos,
		)
	default:
		return "Note enregistrée."
	}
}

func isSkip(body string) bool {
	switch strings.ToLower(strings.TrimSpace(body)) {
	case "non", "passer":
		return true
	default:
		return false
	}
}
