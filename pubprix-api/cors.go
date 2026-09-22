// Middleware CORS — nécessaire pour que l'app Flutter Web (servie sur un
// port de dev qui change à chaque lancement) puisse appeler cette API
// depuis le navigateur.
package main

import (
	"net/http"
	"os"
	"regexp"
)

var localhostOriginPattern = regexp.MustCompile(`^https?://(localhost|127\.0\.0\.1)(:\d+)?$`)

// corsAllowedOrigin renvoie la valeur à utiliser pour
// Access-Control-Allow-Origin pour une requête donnée, ou "" si l'origine
// n'est pas autorisée. Configuré via CORS_ALLOWED_ORIGIN :
//
//   - non définie (défaut dev) : autorise localhost/127.0.0.1 sur
//     n'importe quel port, en reflétant l'Origin de la requête — le port
//     du serveur de dev Flutter change à chaque lancement, une valeur
//     fixe ne conviendrait pas.
//   - "*" : autorise toutes les origines.
//   - une valeur exacte (ex. "https://pubprix.com") : liste blanche
//     stricte à une seule origine. À privilégier en prod une fois un vrai
//     domaine en place (voir aussi la note dans README/ticket sur une
//     liste blanche à plusieurs origines si besoin).
func corsAllowedOrigin(requestOrigin string) string {
	if requestOrigin == "" {
		return ""
	}

	switch configured := os.Getenv("CORS_ALLOWED_ORIGIN"); configured {
	case "":
		if localhostOriginPattern.MatchString(requestOrigin) {
			return requestOrigin
		}
		return ""
	case "*":
		return "*"
	default:
		if configured == requestOrigin {
			return requestOrigin
		}
		return ""
	}
}

// withCORS ajoute les en-têtes CORS et court-circuite les requêtes
// OPTIONS (preflight) avant qu'elles n'atteignent le mux.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowed := corsAllowedOrigin(r.Header.Get("Origin")); allowed != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
			// La valeur d'Access-Control-Allow-Origin dépend de la requête
			// (Origin reflétée) : évite qu'un cache serve cette réponse à
			// une autre origine.
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-Token")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
