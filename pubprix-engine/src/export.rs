//! Export JPEG — RS-05.
//!
//! Vise une taille raisonnable pour les réseaux sociaux (< ~1 Mo) en
//! réduisant progressivement la qualité JPEG si besoin.

use image::codecs::jpeg::JpegEncoder;
use image::{ImageBuffer, Rgb, RgbaImage};

use crate::error::EngineError;

const TARGET_MAX_BYTES: usize = 1_000_000;
const MIN_QUALITY: u8 = 45;
const QUALITY_STEP: u8 = 10;

pub fn encode_jpeg(canvas: &RgbaImage) -> Result<Vec<u8>, EngineError> {
    let rgb: ImageBuffer<Rgb<u8>, Vec<u8>> = ImageBuffer::from_fn(canvas.width(), canvas.height(), |x, y| {
        let p = canvas.get_pixel(x, y);
        Rgb([p[0], p[1], p[2]])
    });

    let mut quality: u8 = 85;
    let mut last_result: Result<Vec<u8>, EngineError> = encode_at_quality(&rgb, quality);

    while let Ok(bytes) = &last_result {
        if bytes.len() <= TARGET_MAX_BYTES || quality <= MIN_QUALITY {
            break;
        }
        quality = quality.saturating_sub(QUALITY_STEP).max(MIN_QUALITY);
        last_result = encode_at_quality(&rgb, quality);
    }

    last_result
}

fn encode_at_quality(rgb: &ImageBuffer<Rgb<u8>, Vec<u8>>, quality: u8) -> Result<Vec<u8>, EngineError> {
    let mut buffer = Vec::new();
    let mut encoder = JpegEncoder::new_with_quality(&mut buffer, quality);
    encoder
        .encode(rgb.as_raw(), rgb.width(), rgb.height(), image::ExtendedColorType::Rgb8)
        .map_err(|e| EngineError::Encode(e.to_string()))?;
    Ok(buffer)
}

#[cfg(test)]
mod tests {
    use super::*;
    use image::Rgba;

    #[test]
    fn encode_jpeg_produces_valid_jpeg_bytes() {
        let canvas = RgbaImage::from_pixel(64, 64, Rgba([120, 60, 200, 255]));
        let bytes = encode_jpeg(&canvas).expect("encode should succeed");
        assert!(bytes.len() > 2);
        assert_eq!(&bytes[0..2], &[0xFF, 0xD8], "un JPEG doit commencer par le marqueur SOI");
    }

    #[test]
    fn encode_jpeg_stays_under_target_size_for_a_noisy_large_image() {
        let mut canvas = RgbaImage::new(1080, 1080);
        for (x, y, pixel) in canvas.enumerate_pixels_mut() {
            *pixel = Rgba([(x % 256) as u8, (y % 256) as u8, ((x + y) % 256) as u8, 255]);
        }
        let bytes = encode_jpeg(&canvas).expect("encode should succeed");
        assert!(bytes.len() <= TARGET_MAX_BYTES, "taille obtenue: {} octets", bytes.len());
    }
}
