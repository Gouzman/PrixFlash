//! Requête/réponse de POST /generate — contrat partagé avec l'API Go.

use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
pub struct GenerateRequest {
    pub photo_urls: Vec<String>,
    pub price: f64,
    pub name: Option<String>,
    #[allow(dead_code)] // format n'influence que la taille du canevas pour l'instant.
    pub format: String, // "whatsapp" | "facebook" | "instagram" | "story"
}

/// `image_base64` contient le JPEG encodé en base64 (voir note dans
/// generate.rs sur ce choix). Ce n'est volontairement pas une URL : c'est
/// à l'API Go de décider où héberger le fichier final (ex: upload vers le
/// même bucket R2 que les photos sources) et de construire l'URL publique.
#[derive(Serialize)]
pub struct GenerateResponse {
    pub status: String, // "ok" | "error"
    #[serde(skip_serializing_if = "Option::is_none")]
    pub image_base64: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub content_type: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

impl GenerateResponse {
    pub fn ok(image_base64: String) -> Self {
        Self {
            status: "ok".to_string(),
            image_base64: Some(image_base64),
            content_type: Some("image/jpeg".to_string()),
            error: None,
        }
    }

    pub fn error(message: impl Into<String>) -> Self {
        Self {
            status: "error".to_string(),
            image_base64: None,
            content_type: None,
            error: Some(message.into()),
        }
    }
}
