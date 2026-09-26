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
