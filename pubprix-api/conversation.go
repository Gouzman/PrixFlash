// Machine à états de conversation WhatsApp par vendeur — WA-03.
//
// Stockage en mémoire pour l'instant (comme les drafts à l'époque de
// Pubprix) : une entrée par numéro de téléphone, à remplacer par une
// vraie base de données avec INF-06.
package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type SellerState string

const (
	StateAttenteNomBoutique SellerState = "ATTENTE_NOM_BOUTIQUE"
	StateAttenteLogo        SellerState = "ATTENTE_LOGO"
	StatePret               SellerState = "PRET"
	StateReceptionPhotos    SellerState = "RECEPTION_PHOTOS"
	StateAttenteStyle       SellerState = "ATTENTE_STYLE"
	StateGeneration         SellerState = "GENERATION"
)

// maxDraftPhotos borne le nombre de photos par publication WhatsApp (1 à
// 6) — distinct de minPhotos/maxPhotos (photos.go), qui bornent l'upload
// multipart de l'ancien flux Flutter (2 à 6).
const maxDraftPhotos = 6

const readyMessage = "Envoie-moi 1 à 6 photos de ton produit, avec le prix si tu veux."

// styleOptions liste les styles de fond proposés à la fin de la
// réception des photos. Pas de vrais boutons interactifs WhatsApp pour
// l'instant (support incertain sur le bac à sable Twilio) : liste
// numérotée en texte, réponse par "1", "2" ou "3", ou par le nom de
// l'option — ticket séparé pour les reply buttons si besoin.
var styleOptions = []string{"Fond blanc", "Fond couleur", "Mise en scène"}

var styleChoicePrompt = buildStyleChoicePrompt()

func buildStyleChoicePrompt() string {
	var b strings.Builder
	b.WriteString("Quel style de fond pour ta publication ?\n")
	for i, opt := range styleOptions {
		fmt.Fprintf(&b, "%d. %s\n", i+1, opt)
	}
	b.WriteString("Réponds avec le numéro de ton choix.")
	return b.String()
}

const generationStubMessage = "Photos reçues, traitement en cours de développement, merci de ta patience"

// SellerDraft est la publication en préparation pour un vendeur : juste
// les références des photos reçues, les notes texte qui les accompagnent
// (prix, détails...) et le style choisi — pas de traitement d'image à ce
// stade (tickets IMG à venir).
type SellerDraft struct {
	Photos []string
	Notes  []string
	Style  string
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
	case StateAttenteStyle:
		return handleAttenteStyle(seller, body)
	case StateGeneration:
		// Ne devrait pas être observé en pratique : la transition vers
		// GENERATION se résout de façon synchrone dans le même appel (voir
		// handleGeneration). Filet de sécurité si un message arrive quand
		// même pendant ce court état.
		return "Ta publication est en cours de traitement, merci de patienter."
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

	if !hasMedia && isFinishKeyword(body) {
		if len(seller.Draft.Photos) == 0 {
			return "Il me faut au moins une photo avant de continuer. Envoie une photo de ton produit."
		}
		seller.State = StateAttenteStyle
		return styleChoicePrompt
	}

	if !hasMedia && body == "" {
		return "Envoie une photo ou un prix pour continuer ta publication."
	}
	return addToDraft(seller.Draft, body, mediaRefs)
}

// handleAttenteStyle traite la réponse au choix de style (numéro 1/2/3 ou
// nom de l'option). Une réponse valide enregistre le style sur le
// brouillon et fait passer le vendeur en GENERATION — état qui se résout
// immédiatement pour l'instant (voir handleGeneration).
func handleAttenteStyle(seller *Seller, body string) string {
	choice, ok := matchStyleChoice(body)
	if !ok {
		return "Je n'ai pas compris ton choix.\n\n" + styleChoicePrompt
	}
	if seller.Draft == nil {
		// Filet de sécurité : ne devrait pas arriver, on vient forcément
		// de RECEPTION_PHOTOS avec un brouillon non vide.
		seller.Draft = &SellerDraft{}
	}
	seller.Draft.Style = choice
	seller.State = StateGeneration
	return handleGeneration(seller)
}

// handleGeneration représente le comportement de l'état GENERATION.
// Stub pour l'instant : le traitement d'image n'est pas encore branché
// (tickets IMG à venir) — on répond immédiatement et on remet le vendeur
// en PRET pour qu'il puisse démarrer un nouveau produit.
func handleGeneration(seller *Seller) string {
	seller.State = StatePret
	seller.Draft = nil
	return generationStubMessage
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

// finishKeywords, une fois chaque entrée passée par normalizeText.
// L'apostrophe de "c'est fini" devient un espace lors de la
// normalisation, d'où "c est fini" plutôt que "cest fini".
var finishKeywords = map[string]bool{
	"c est fini": true,
	"fini":       true,
	"termine":    true,
	"go":         true,
}

// isFinishKeyword détecte un signal de fin d'envoi de photos, insensible
// à la casse, aux accents et à la ponctuation environnante (guillemets
// exacts non requis).
func isFinishKeyword(body string) bool {
	return finishKeywords[normalizeText(body)]
}

// matchStyleChoice reconnaît une réponse au choix de style : un chiffre
// (1/2/3) ou le nom de l'option (insensible à la casse/accents).
func matchStyleChoice(body string) (string, bool) {
	trimmed := strings.TrimSpace(body)
	if idx, err := strconv.Atoi(trimmed); err == nil && idx >= 1 && idx <= len(styleOptions) {
		return styleOptions[idx-1], true
	}

	normalized := normalizeText(body)
	for _, opt := range styleOptions {
		if normalizeText(opt) == normalized {
			return opt, true
		}
	}
	return "", false
}

var accentReplacer = strings.NewReplacer(
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"à", "a", "â", "a", "ä", "a",
	"î", "i", "ï", "i",
	"ô", "o", "ö", "o",
	"ù", "u", "û", "u", "ü", "u",
	"ç", "c",
)

// normalizeText met en minuscules, retire les accents et remplace toute
// ponctuation par des espaces (compressés), pour comparer du texte libre
// aux mots-clés attendus sans exiger une saisie exacte.
func normalizeText(s string) string {
	s = accentReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))

	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
