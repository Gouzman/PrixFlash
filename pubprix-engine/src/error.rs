//! Erreurs du pipeline de génération — jamais de panique en réponse à une
//! entrée invalide ou une dépendance externe défaillante : tout se
//! transforme en réponse HTTP avec un statut adapté.

use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use std::fmt;

use crate::model::GenerateResponse;

#[derive(Debug)]
pub enum EngineError {
    NoPhotos,
    InvalidPrice,
    Download(String),
    Decode(String),
    Encode(String),
}

impl fmt::Display for EngineError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            EngineError::NoPhotos => write!(f, "aucune photo fournie"),
            EngineError::InvalidPrice => write!(f, "prix invalide"),
            EngineError::Download(url) => write!(f, "téléchargement impossible: {url}"),
            EngineError::Decode(msg) => write!(f, "décodage image impossible: {msg}"),
            EngineError::Encode(msg) => write!(f, "export JPEG impossible: {msg}"),
        }
    }
}

impl std::error::Error for EngineError {}

impl IntoResponse for EngineError {
    fn into_response(self) -> Response {
        let status = match self {
            EngineError::NoPhotos | EngineError::InvalidPrice => StatusCode::BAD_REQUEST,
            EngineError::Download(_) | EngineError::Decode(_) => StatusCode::BAD_GATEWAY,
            EngineError::Encode(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };
        (status, Json(GenerateResponse::error(self.to_string()))).into_response()
    }
}
