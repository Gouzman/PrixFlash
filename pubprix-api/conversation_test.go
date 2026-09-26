package main

import (
	"fmt"
	"strings"
	"testing"
)

// uniquePhone évite les collisions entre tests sur la map globale sellers.
var phoneCounter int

func uniquePhone(t *testing.T) string {
	t.Helper()
	phoneCounter++
	return fmt.Sprintf("whatsapp:+2250700%06d", phoneCounter)
}

func getSeller(t *testing.T, phone string) *Seller {
	t.Helper()
	sellersMu.Lock()
	defer sellersMu.Unlock()
	s, ok := sellers[phone]
	if !ok {
		t.Fatalf("aucun vendeur enregistré pour %s", phone)
	}
	return s
}

func TestUnknownNumberStartsOnboarding(t *testing.T) {
	phone := uniquePhone(t)

	reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: "Bonjour"})

	if !strings.Contains(reply, "boutique") {
		t.Errorf("réponse d'accueil attendue avec une demande de nom de boutique, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StateAttenteNomBoutique {
		t.Errorf("état attendu %s, obtenu %s", StateAttenteNomBoutique, seller.State)
	}
}

func TestFirstMessageContentIsIgnored(t *testing.T) {
	phone := uniquePhone(t)

	// Le tout premier message ne doit pas être interprété comme un nom de
	// boutique, même s'il contient du texte.
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ceci ne doit pas être pris pour un nom de boutique"})

	seller := getSeller(t, phone)
	if seller.ShopName != "" {
		t.Errorf("ShopName devrait être vide après le premier message, obtenu %q", seller.ShopName)
	}
	if seller.State != StateAttenteNomBoutique {
		t.Errorf("état attendu %s, obtenu %s", StateAttenteNomBoutique, seller.State)
	}
}

func TestShopNameThenSkipLogo(t *testing.T) {
	phone := uniquePhone(t)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})

	reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: "Chaussures Awa"})
	if !strings.Contains(reply, "logo") {
		t.Errorf("réponse attendue avec une question sur le logo, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.ShopName != "Chaussures Awa" {
		t.Errorf("ShopName attendu %q, obtenu %q", "Chaussures Awa", seller.ShopName)
	}
	if seller.State != StateAttenteLogo {
		t.Errorf("état attendu %s, obtenu %s", StateAttenteLogo, seller.State)
	}

	for _, skipWord := range []string{"non", "NON", "passer", "Passer"} {
		phone2 := uniquePhone(t)
		HandleIncomingMessage(IncomingMessage{From: phone2, Body: "salut"})
		HandleIncomingMessage(IncomingMessage{From: phone2, Body: "Boutique X"})
		reply := HandleIncomingMessage(IncomingMessage{From: phone2, Body: skipWord})
		if !strings.Contains(reply, "photos") {
			t.Errorf("skipWord=%q: réponse attendue avec le message prêt, reçu: %q", skipWord, reply)
		}
		s := getSeller(t, phone2)
		if s.State != StatePret {
			t.Errorf("skipWord=%q: état attendu %s, obtenu %s", skipWord, StatePret, s.State)
		}
		if s.LogoRef != "" {
			t.Errorf("skipWord=%q: LogoRef devrait rester vide, obtenu %q", skipWord, s.LogoRef)
		}
	}
}

func TestLogoImageSavesReference(t *testing.T) {
	phone := uniquePhone(t)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ma Boutique"})

	reply := HandleIncomingMessage(IncomingMessage{
		From:      phone,
		MediaRefs: []string{"https://twilio.example/media/logo.jpg"},
	})

	if !strings.Contains(reply, "photos") {
		t.Errorf("réponse attendue avec le message prêt après le logo, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.LogoRef != "https://twilio.example/media/logo.jpg" {
		t.Errorf("LogoRef attendu la référence du média, obtenu %q", seller.LogoRef)
	}
	if seller.State != StatePret {
		t.Errorf("état attendu %s, obtenu %s", StatePret, seller.State)
	}
}

func TestLogoStepRepromptsOnUnrelatedText(t *testing.T) {
	phone := uniquePhone(t)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ma Boutique"})

	reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: "peut-être plus tard"})
	if !strings.Contains(reply, "logo") {
		t.Errorf("réponse attendue redemandant le logo, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StateAttenteLogo {
		t.Errorf("état attendu %s (inchangé), obtenu %s", StateAttenteLogo, seller.State)
	}
}

func TestPhotoFromPretStartsDraftAndMovesToReceptionPhotos(t *testing.T) {
	phone := uniquePhone(t)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ma Boutique"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "non"})

	reply := HandleIncomingMessage(IncomingMessage{
		From:      phone,
		Body:      "15000 FCFA",
		MediaRefs: []string{"https://twilio.example/media/photo1.jpg"},
	})

	if !strings.Contains(reply, "1/6") {
		t.Errorf("réponse attendue avec le compteur 1/6, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StateReceptionPhotos {
		t.Errorf("état attendu %s, obtenu %s", StateReceptionPhotos, seller.State)
	}
	if seller.Draft == nil {
		t.Fatal("un brouillon devrait avoir été créé")
	}
	if len(seller.Draft.Photos) != 1 || seller.Draft.Photos[0] != "https://twilio.example/media/photo1.jpg" {
		t.Errorf("photos du brouillon inattendues: %v", seller.Draft.Photos)
	}
	if len(seller.Draft.Notes) != 1 || seller.Draft.Notes[0] != "15000 FCFA" {
		t.Errorf("notes du brouillon inattendues: %v", seller.Draft.Notes)
	}
}

func TestReceptionPhotosAccumulatesAcrossMessages(t *testing.T) {
	phone := uniquePhone(t)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ma Boutique"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "non"})
	HandleIncomingMessage(IncomingMessage{From: phone, MediaRefs: []string{"photo1"}})
	HandleIncomingMessage(IncomingMessage{From: phone, MediaRefs: []string{"photo2"}})
	reply := HandleIncomingMessage(IncomingMessage{From: phone, MediaRefs: []string{"photo3"}, Body: "encore une vue"})

	if !strings.Contains(reply, "3/6") {
		t.Errorf("réponse attendue avec le compteur 3/6, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if len(seller.Draft.Photos) != 3 {
		t.Errorf("3 photos attendues, obtenu %d: %v", len(seller.Draft.Photos), seller.Draft.Photos)
	}
	if len(seller.Draft.Notes) != 1 {
		t.Errorf("1 note attendue, obtenu %d: %v", len(seller.Draft.Notes), seller.Draft.Notes)
	}
}

func TestReceptionPhotosEnforcesMaxSix(t *testing.T) {
	phone := uniquePhone(t)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ma Boutique"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "non"})
	for i := 0; i < 6; i++ {
		HandleIncomingMessage(IncomingMessage{From: phone, MediaRefs: []string{fmt.Sprintf("photo%d", i)}})
	}

	reply := HandleIncomingMessage(IncomingMessage{From: phone, MediaRefs: []string{"photo-trop"}})
	if !strings.Contains(reply, "maximum") {
		t.Errorf("réponse attendue signalant le maximum atteint, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if len(seller.Draft.Photos) != 6 {
		t.Errorf("le brouillon ne devrait pas dépasser 6 photos, obtenu %d", len(seller.Draft.Photos))
	}
}

func TestKnownSellerSendingPhotoSkipsOnboarding(t *testing.T) {
	phone := uniquePhone(t)
	// Onboarding complet une première fois.
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ma Boutique"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "non"})
	// Le vendeur est maintenant en PRET ; il revient plus tard avec une
	// photo directement, sans repasser par l'accueil.
	reply := HandleIncomingMessage(IncomingMessage{From: phone, MediaRefs: []string{"photo-retour"}})

	if !strings.Contains(reply, "1/6") {
		t.Errorf("réponse attendue relançant la réception de photos, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StateReceptionPhotos {
		t.Errorf("état attendu %s, obtenu %s", StateReceptionPhotos, seller.State)
	}
	if seller.ShopName != "Ma Boutique" {
		t.Errorf("le nom de boutique enregistré ne doit pas être perdu, obtenu %q", seller.ShopName)
	}
}

func TestTwoSellersDoNotInterfere(t *testing.T) {
	phoneA := uniquePhone(t)
	phoneB := uniquePhone(t)

	HandleIncomingMessage(IncomingMessage{From: phoneA, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phoneA, Body: "Boutique A"})

	HandleIncomingMessage(IncomingMessage{From: phoneB, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phoneB, Body: "Boutique B"})

	sellerA := getSeller(t, phoneA)
	sellerB := getSeller(t, phoneB)

	if sellerA.ShopName != "Boutique A" {
		t.Errorf("vendeur A: nom attendu %q, obtenu %q", "Boutique A", sellerA.ShopName)
	}
	if sellerB.ShopName != "Boutique B" {
		t.Errorf("vendeur B: nom attendu %q, obtenu %q", "Boutique B", sellerB.ShopName)
	}
}

// onboardToReceptionPhotos amène un vendeur jusqu'en RECEPTION_PHOTOS
// avec nPhotos déjà envoyées, pour les tests portant sur la suite du
// flux (signal de fin, choix de style).
func onboardToReceptionPhotos(t *testing.T, phone string, nPhotos int) {
	t.Helper()
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "salut"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "Ma Boutique"})
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "non"})
	for i := 0; i < nPhotos; i++ {
		HandleIncomingMessage(IncomingMessage{From: phone, MediaRefs: []string{fmt.Sprintf("photo%d", i)}})
	}
}

func TestFinishKeywordWithoutPhotosStaysInReceptionPhotos(t *testing.T) {
	phone := uniquePhone(t)
	// Arriver en RECEPTION_PHOTOS ajoute toujours au moins une photo (voir
	// handlePret), donc "brouillon vide en RECEPTION_PHOTOS" n'est pas
	// atteignable via le flux normal. On force l'état directement pour
	// vérifier que le garde-fou du ticket reste correct si jamais ça
	// arrive (ex: évolution future du flux).
	sellersMu.Lock()
	sellers[phone] = &Seller{
		PhoneNumber: phone,
		State:       StateReceptionPhotos,
		ShopName:    "Ma Boutique",
		Draft:       &SellerDraft{},
	}
	sellersMu.Unlock()

	reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: "c'est fini"})

	if !strings.Contains(reply, "au moins une photo") {
		t.Errorf("réponse attendue demandant au moins une photo, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StateReceptionPhotos {
		t.Errorf("état attendu %s (inchangé), obtenu %s", StateReceptionPhotos, seller.State)
	}
}

func TestFinishKeywordVariantsMoveToAttenteStyle(t *testing.T) {
	variants := []string{"c'est fini", "C'EST FINI", "fini", "Fini.", "terminé", "TERMINÉ", "go", "Go !"}

	for _, kw := range variants {
		phone := uniquePhone(t)
		onboardToReceptionPhotos(t, phone, 2)

		reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: kw})

		if !strings.Contains(reply, "1.") || !strings.Contains(reply, "Fond blanc") {
			t.Errorf("mot-clé %q: réponse attendue listant les styles, reçu: %q", kw, reply)
		}
		seller := getSeller(t, phone)
		if seller.State != StateAttenteStyle {
			t.Errorf("mot-clé %q: état attendu %s, obtenu %s", kw, StateAttenteStyle, seller.State)
		}
		if len(seller.Draft.Photos) != 2 {
			t.Errorf("mot-clé %q: les photos du brouillon ne devraient pas être perdues, obtenu %d", kw, len(seller.Draft.Photos))
		}
	}
}

func TestNonFinishTextStaysInReceptionPhotos(t *testing.T) {
	phone := uniquePhone(t)
	onboardToReceptionPhotos(t, phone, 1)

	// "fini" est un mot-clé, mais une phrase qui le contient sans être
	// exactement ce mot-clé (normalisé) ne doit pas déclencher la fin.
	reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: "je n'ai pas fini de choisir le prix"})

	if strings.Contains(reply, "Fond blanc") {
		t.Errorf("ne devrait pas déclencher le choix de style, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StateReceptionPhotos {
		t.Errorf("état attendu %s (inchangé), obtenu %s", StateReceptionPhotos, seller.State)
	}
	if len(seller.Draft.Notes) != 1 {
		t.Errorf("le texte devrait être enregistré comme note, obtenu %v", seller.Draft.Notes)
	}
}

func TestStyleChoiceByNumberCompletesGenerationStubAndResetsToPret(t *testing.T) {
	for i := range styleOptions {
		phone := uniquePhone(t)
		onboardToReceptionPhotos(t, phone, 2)
		HandleIncomingMessage(IncomingMessage{From: phone, Body: "fini"})

		reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: fmt.Sprintf("%d", i+1)})

		if reply != generationStubMessage {
			t.Errorf("option %d: réponse attendue le message stub de génération, reçu: %q", i+1, reply)
		}
		seller := getSeller(t, phone)
		if seller.State != StatePret {
			t.Errorf("option %d: état attendu %s après le stub, obtenu %s", i+1, StatePret, seller.State)
		}
		if seller.Draft != nil {
			t.Errorf("option %d: le brouillon devrait être remis à zéro, obtenu %+v", i+1, seller.Draft)
		}
	}
}

func TestStyleChoiceIsActuallyWrittenOntoDraftBeforeStubClearsIt(t *testing.T) {
	phone := uniquePhone(t)
	onboardToReceptionPhotos(t, phone, 2)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "fini"})

	// Capture le pointeur vers le brouillon avant le choix de style : le
	// stub de GENERATION remet seller.Draft à nil, mais l'objet pointé
	// par draftBefore reste inspectable.
	draftBefore := getSeller(t, phone).Draft
	if draftBefore == nil {
		t.Fatal("le brouillon devrait exister avant le choix de style")
	}

	HandleIncomingMessage(IncomingMessage{From: phone, Body: "2"}) // "Fond couleur"

	if draftBefore.Style != "Fond couleur" {
		t.Errorf("le style choisi devrait avoir été écrit sur le brouillon, obtenu %q", draftBefore.Style)
	}
}

func TestStyleChoiceByNameIsAccepted(t *testing.T) {
	phone := uniquePhone(t)
	onboardToReceptionPhotos(t, phone, 1)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "fini"})

	reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: "fond couleur"})

	if reply != generationStubMessage {
		t.Errorf("réponse attendue le message stub de génération, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StatePret {
		t.Errorf("état attendu %s, obtenu %s", StatePret, seller.State)
	}
}

func TestStyleChoiceInvalidRepromptsWithoutLosingDraft(t *testing.T) {
	phone := uniquePhone(t)
	onboardToReceptionPhotos(t, phone, 2)
	HandleIncomingMessage(IncomingMessage{From: phone, Body: "fini"})

	reply := HandleIncomingMessage(IncomingMessage{From: phone, Body: "je sais pas"})

	if !strings.Contains(reply, "Fond blanc") {
		t.Errorf("réponse attendue relistant les options, reçu: %q", reply)
	}
	seller := getSeller(t, phone)
	if seller.State != StateAttenteStyle {
		t.Errorf("état attendu %s (inchangé), obtenu %s", StateAttenteStyle, seller.State)
	}
	if seller.Draft == nil || len(seller.Draft.Photos) != 2 {
		t.Errorf("le brouillon ne devrait pas être perdu sur un choix invalide")
	}
}
