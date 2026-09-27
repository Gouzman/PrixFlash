# Pubprix — microservice de détourage de fond (IMG-01).
#
# Ne fait QUE le détourage : reçoit une image, renvoie un PNG à fond
# transparent (canal alpha). Pas de composition, pas d'overlay, pas de
# choix de couleur de fond — ça, c'est le moteur Rust (IMG-02, ticket
# séparé).
from __future__ import annotations

import io
import logging
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import Response
from PIL import Image, UnidentifiedImageError
from rembg import new_session, remove

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("pubprix-detourage")

# u2net : modèle historique de rembg, ~176 Mo, largement éprouvé. D'autres
# versions de rembg pointent par défaut vers des modèles bien plus gros
# (ex. bria-rmbg-2.0, ~1 Go) dont le téléchargement s'est révélé peu fiable
# ici — configurable si besoin via REMBG_MODEL plutôt qu'en dur.
MODEL_NAME = os.getenv("REMBG_MODEL", "u2net")

_session = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global _session
    logger.info("chargement du modèle de détourage (%s)…", MODEL_NAME)
    _session = new_session(MODEL_NAME)
    logger.info("modèle de détourage prêt.")
    yield


app = FastAPI(title="Pubprix — détourage", lifespan=lifespan)

# Formats acceptés en entrée. rembg/Pillow gèrent bien d'autres formats,
# mais on aligne la liste sur celle déjà validée côté API Go (photos.go)
# pour rester cohérent sur toute la chaîne.
ALLOWED_CONTENT_TYPES = {"image/jpeg", "image/png", "image/webp"}

MAX_UPLOAD_BYTES = 20 * 1024 * 1024  # 20 Mo, généreux pour une photo produit


@app.get("/healthz")
def healthz() -> str:
    return "ok"


@app.post("/remove-background")
async def remove_background(file: UploadFile = File(...)) -> Response:
    if file.content_type not in ALLOWED_CONTENT_TYPES:
        raise HTTPException(
            status_code=400,
            detail=f"type de fichier non supporté: {file.content_type}",
        )

    raw = await file.read()
    if not raw:
        raise HTTPException(status_code=400, detail="fichier vide")
    if len(raw) > MAX_UPLOAD_BYTES:
        raise HTTPException(
            status_code=400,
            detail=f"fichier trop volumineux ({len(raw)} octets, max {MAX_UPLOAD_BYTES})",
        )

    # Valide que c'est bien une image décodable avant de la passer à
    # rembg — évite de faire planter le modèle sur des octets invalides
    # et donne un message d'erreur clair (400) plutôt qu'une 500 opaque.
    try:
        Image.open(io.BytesIO(raw)).verify()
    except (UnidentifiedImageError, OSError, SyntaxError, ValueError) as exc:
        raise HTTPException(status_code=400, detail="image invalide ou corrompue") from exc

    try:
        result = remove(raw, session=_session)
    except Exception as exc:  # défense en profondeur : rembg ne doit jamais faire planter le service
        logger.exception("échec du détourage")
        raise HTTPException(status_code=500, detail="échec du traitement de l'image") from exc

    return Response(content=result, media_type="image/png")
