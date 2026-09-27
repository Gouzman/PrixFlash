//! Orchestration du pipeline complet : téléchargement -> composition ->
//! overlay -> export -> encodage base64. Appelé par le handler axum de
//! main.rs.

use axum::extract::State;
use axum::response::IntoResponse;
use axum::Json;
use base64::engine::general_purpose::STANDARD as BASE64;
use base64::Engine as _;

use crate::background::{flatten_on_style_background, BackgroundStyle};
use crate::compose::compose_canvas;
use crate::download::fetch_all;
use crate::error::EngineError;
use crate::export::encode_jpeg;
use crate::model::{GenerateRequest, GenerateResponse};
use crate::overlay::draw_price_and_name;
use image::DynamicImage;

#[derive(Clone)]
pub struct AppState {
    pub http_client: reqwest::Client,
}

fn canvas_size_for_format(format: &str) -> (u32, u32) {
    match format {
        "story" => (1080, 1920),
        _ => (1080, 1080), // whatsapp / facebook / instagram : carré, volontairement simple.
    }
}

fn format_price(price: f64, currency: &str) -> String {
    let rounded = price.round().max(0.0) as i64;
    let digits: Vec<char> = rounded.to_string().chars().collect();
    let mut grouped = String::new();
    for (i, c) in digits.iter().enumerate() {
        if i != 0 && (digits.len() - i) % 3 == 0 {
            grouped.push(' ');
        }
        grouped.push(*c);
    }
    format!("{grouped} {currency}")
}

pub async fn run_generate(client: &reqwest::Client, req: &GenerateRequest) -> Result<Vec<u8>, EngineError> {
    if req.photo_urls.is_empty() {
        return Err(EngineError::NoPhotos);
    }
    // 0 = pas de prix fourni (flux WhatsApp) ; seule une valeur négative
    // ou non finie est une véritable erreur d'entrée.
    if !req.price.is_finite() || req.price < 0.0 {
        return Err(EngineError::InvalidPrice);
    }

    let mut photos = fetch_all(client, &req.photo_urls).await?;

    // Photos détourées (canal alpha) : on pose le fond du style choisi
    // derrière chacune avant la composition multi-angles, qui suppose des
    // images déjà opaques. Sans style (ancien flux), comportement inchangé.
    if let Some(style) = req.style.as_deref().and_then(BackgroundStyle::parse) {
        photos = photos
            .iter()
            .map(|photo| DynamicImage::ImageRgba8(flatten_on_style_background(photo, style)))
            .collect();
    }

    let (canvas_w, canvas_h) = canvas_size_for_format(&req.format);
    let canvas = compose_canvas(&photos, canvas_w, canvas_h);

    // Pas encore de texte de vente généré (prix/nom) pour le flux
    // WhatsApp à ce stade : l'overlay ne se déclenche que si l'appelant a
    // effectivement fourni un prix (draw_price_and_name se protège aussi
    // elle-même si jamais les deux sont vides).
    let overlaid = if req.price > 0.0 || req.name.is_some() {
        let price_label = if req.price > 0.0 { format_price(req.price, "FCFA") } else { String::new() };
        draw_price_and_name(&canvas, &price_label, req.name.as_deref())
    } else {
        canvas
    };

    encode_jpeg(&overlaid)
}

pub async fn generate_handler(
    State(state): State<AppState>,
    Json(payload): Json<GenerateRequest>,
) -> impl IntoResponse {
    match run_generate(&state.http_client, &payload).await {
        Ok(jpeg_bytes) => Json(GenerateResponse::ok(BASE64.encode(jpeg_bytes))).into_response(),
        Err(err) => err.into_response(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn format_price_groups_thousands() {
        assert_eq!(format_price(12500.0, "FCFA"), "12 500 FCFA");
        assert_eq!(format_price(500.0, "FCFA"), "500 FCFA");
        assert_eq!(format_price(1_250_000.0, "FCFA"), "1 250 000 FCFA");
    }

    #[test]
    fn canvas_size_depends_on_format() {
        assert_eq!(canvas_size_for_format("story"), (1080, 1920));
        assert_eq!(canvas_size_for_format("whatsapp"), (1080, 1080));
        assert_eq!(canvas_size_for_format("unknown"), (1080, 1080));
    }
}
