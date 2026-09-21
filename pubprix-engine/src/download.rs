//! Téléchargement des photos sources — RS-02.
//!
//! Garantie : on ne fait que des GET vers les URLs fournies et on ne
//! décode qu'en mémoire (`image::DynamicImage`). Les octets originaux
//! téléchargés ne sont jamais réécrits ni renvoyés vers leur source ; le
//! pipeline en aval (compose.rs, overlay.rs) ne travaille que sur des
//! copies/dérivés en mémoire de ces images décodées.

use image::DynamicImage;

use crate::error::EngineError;

pub async fn fetch_photo(client: &reqwest::Client, url: &str) -> Result<DynamicImage, EngineError> {
    let response = client
        .get(url)
        .send()
        .await
        .map_err(|_| EngineError::Download(url.to_string()))?;

    if !response.status().is_success() {
        return Err(EngineError::Download(url.to_string()));
    }

    let bytes = response
        .bytes()
        .await
        .map_err(|_| EngineError::Download(url.to_string()))?;

    image::load_from_memory(&bytes).map_err(|e| EngineError::Decode(e.to_string()))
}

pub async fn fetch_all(client: &reqwest::Client, urls: &[String]) -> Result<Vec<DynamicImage>, EngineError> {
    let mut photos = Vec::with_capacity(urls.len());
    for url in urls {
        photos.push(fetch_photo(client, url).await?);
    }
    Ok(photos)
}
