# pubprix-detourage

Microservice de détourage de fond (IMG-01). Reçoit une photo produit,
renvoie un PNG à fond transparent. Ne fait **que** ça : pas de
composition, pas d'overlay, pas de choix de couleur de fond — ce sera le
rôle du moteur Rust (`pubprix-engine`, IMG-02).

## Installer

Nécessite Python 3.10+ (testé avec Python 3.14).

```bash
python -m venv .venv
.venv\Scripts\activate        # Windows
# source .venv/bin/activate   # macOS/Linux

pip install -r requirements.txt
```

## Lancer

```bash
uvicorn main:app --port 8082
```

Ports des autres services du projet : API Go = 8080, moteur Rust = 8081,
ce service = 8082.

### Téléchargement du modèle au premier lancement

Au premier démarrage, le service télécharge le modèle de détourage
(`u2net`, ~176 Mo) depuis GitHub et le met en cache dans
`~/.rembg/models/`. Compter 20-30 secondes selon la connexion ; les
démarrages suivants sont quasi instantanés (modèle déjà en cache).

Le modèle est configurable via la variable d'environnement
`REMBG_MODEL` si besoin (voir la liste des modèles disponibles dans
[la documentation de rembg](https://github.com/danielgatis/rembg)) —
`u2net` a été choisi comme valeur par défaut plutôt que le modèle par
défaut de rembg (`bria-rmbg-2.0`, ~1 Go) dont le téléchargement s'est
révélé peu fiable en pratique (interruptions réseau sur un fichier aussi
gros) ; `u2net` est plus petit, plus rapide à charger, et donne de bons
résultats sur des photos produit.

## Endpoints

- `GET /healthz` — vérification de disponibilité.
- `POST /remove-background` — upload multipart (`file`), renvoie
  l'image en PNG (fond transparent). Types acceptés : `image/jpeg`,
  `image/png`, `image/webp`. 20 Mo max.

Exemple :

```bash
curl -X POST http://localhost:8082/remove-background \
  -F "file=@photo.jpg;type=image/jpeg" \
  -o resultat.png
```

## Erreurs

- `400` — fichier absent/vide, type non supporté, ou image
  invalide/corrompue.
- `500` — échec du traitement (modèle non chargé, erreur interne).

Le service ne plante jamais sur une entrée invalide : chaque cas
d'erreur est intercepté et renvoyé avec un code HTTP approprié.
