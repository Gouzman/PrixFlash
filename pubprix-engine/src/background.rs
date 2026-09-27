//! Fond de style derrière une photo détourée — IMG-02.
//!
//! Les photos qui arrivent avec un `style` viennent du microservice de
//! détourage (`pubprix-detourage`) : elles ont un canal alpha, le produit
//! sur fond transparent. Avant la composition multi-angles existante
//! (compose.rs, qui suppose des images opaques), chaque photo est
//! aplatie sur un fond correspondant au style choisi par le vendeur.
//!
//! Pas de vrai décor pour "mise en scène" à ce stade (dégradé simple en
//! attendant) — voir le ticket produit pour la suite.

use image::{imageops, DynamicImage, Rgba, RgbaImage};

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum BackgroundStyle {
    White,
    Color,
    Scene,
}

impl BackgroundStyle {
    /// Reconnaît la valeur envoyée par l'API Go ("white" | "color" |
    /// "scene"). `None` si absente/inconnue — l'appelant garde alors le
    /// comportement historique (photos déjà opaques, pas de fond à
    /// poser).
    pub fn parse(s: &str) -> Option<Self> {
        match s {
            "white" => Some(Self::White),
            "color" => Some(Self::Color),
            "scene" => Some(Self::Scene),
            _ => None,
        }
    }
}

/// Aplatit `photo` (potentiellement avec canal alpha) sur un fond uni ou
/// dégradé correspondant au style. Une image déjà opaque n'est pas
/// affectée visuellement (aucune zone transparente à combler), donc la
/// fonction est sûre à appeler même sur d'anciennes photos non détourées.
pub fn flatten_on_style_background(photo: &DynamicImage, style: BackgroundStyle) -> RgbaImage {
    let (w, h) = image::GenericImageView::dimensions(photo);
    let mut bg = match style {
        BackgroundStyle::White => RgbaImage::from_pixel(w, h, Rgba([255, 255, 255, 255])),
        // Couleur neutre arbitraire pour ce sprint — un vrai choix de
        // couleur (ex. à partir de la marque du vendeur) viendra plus
        // tard.
        BackgroundStyle::Color => RgbaImage::from_pixel(w, h, Rgba([230, 214, 194, 255])),
        BackgroundStyle::Scene => vertical_gradient(w, h, Rgba([225, 225, 232, 255]), Rgba([180, 182, 196, 255])),
    };
    imageops::overlay(&mut bg, photo, 0, 0);
    bg
}

fn vertical_gradient(w: u32, h: u32, top: Rgba<u8>, bottom: Rgba<u8>) -> RgbaImage {
    RgbaImage::from_fn(w, h, |_, y| {
        let t = if h <= 1 { 0.0 } else { y as f32 / (h - 1) as f32 };
        Rgba([
            lerp_u8(top[0], bottom[0], t),
            lerp_u8(top[1], bottom[1], t),
            lerp_u8(top[2], bottom[2], t),
            255,
        ])
    })
}

fn lerp_u8(a: u8, b: u8, t: f32) -> u8 {
    (a as f32 + (b as f32 - a as f32) * t).round() as u8
}

#[cfg(test)]
mod tests {
    use super::*;
    use image::Rgba as PixelRgba;

    fn transparent_square(size: u32) -> DynamicImage {
        DynamicImage::ImageRgba8(RgbaImage::from_pixel(size, size, PixelRgba([0, 0, 0, 0])))
    }

    fn half_opaque_square(size: u32, color: [u8; 4]) -> DynamicImage {
        let mut img = RgbaImage::from_pixel(size, size, PixelRgba([0, 0, 0, 0]));
        for y in 0..size {
            for x in 0..size / 2 {
                img.put_pixel(x, y, PixelRgba(color));
            }
        }
        DynamicImage::ImageRgba8(img)
    }

    #[test]
    fn white_style_fills_transparent_pixels_with_white() {
        let photo = transparent_square(10);
        let result = flatten_on_style_background(&photo, BackgroundStyle::White);
        assert_eq!(*result.get_pixel(0, 0), PixelRgba([255, 255, 255, 255]));
        assert_eq!(*result.get_pixel(9, 9), PixelRgba([255, 255, 255, 255]));
    }

    #[test]
    fn opaque_pixels_are_preserved_over_the_background() {
        let photo = half_opaque_square(10, [10, 20, 30, 255]);
        let result = flatten_on_style_background(&photo, BackgroundStyle::White);
        assert_eq!(*result.get_pixel(0, 5), PixelRgba([10, 20, 30, 255]), "le produit doit rester intact");
        assert_eq!(*result.get_pixel(9, 5), PixelRgba([255, 255, 255, 255]), "le fond doit apparaitre cote transparent");
    }

    #[test]
    fn color_style_uses_a_solid_non_white_background() {
        let photo = transparent_square(4);
        let result = flatten_on_style_background(&photo, BackgroundStyle::Color);
        let pixel = *result.get_pixel(0, 0);
        assert_ne!(pixel, PixelRgba([255, 255, 255, 255]));
        assert_eq!(pixel[3], 255, "le fond doit etre completement opaque");
    }

    #[test]
    fn scene_style_produces_a_vertical_gradient() {
        let photo = transparent_square(1);
        let result = flatten_on_style_background(&photo, BackgroundStyle::Scene);
        // Un dégradé vertical sur une image de hauteur 1 n'a qu'une seule
        // ligne : on vérifie plutôt directement le dégradé brut.
        let tall = vertical_gradient(2, 50, Rgba([225, 225, 232, 255]), Rgba([180, 182, 196, 255]));
        let top = *tall.get_pixel(0, 0);
        let bottom = *tall.get_pixel(0, 49);
        assert_ne!(top, bottom, "le degrade doit varier entre le haut et le bas");
        assert_eq!(result.dimensions(), (1, 1));
    }

    #[test]
    fn output_is_always_fully_opaque() {
        for style in [BackgroundStyle::White, BackgroundStyle::Color, BackgroundStyle::Scene] {
            let photo = half_opaque_square(6, [1, 2, 3, 255]);
            let result = flatten_on_style_background(&photo, style);
            for pixel in result.pixels() {
                assert_eq!(pixel[3], 255, "style {style:?}: tous les pixels doivent etre opaques");
            }
        }
    }

    #[test]
    fn parse_recognizes_known_styles_and_rejects_unknown() {
        assert_eq!(BackgroundStyle::parse("white"), Some(BackgroundStyle::White));
        assert_eq!(BackgroundStyle::parse("color"), Some(BackgroundStyle::Color));
        assert_eq!(BackgroundStyle::parse("scene"), Some(BackgroundStyle::Scene));
        assert_eq!(BackgroundStyle::parse(""), None);
        assert_eq!(BackgroundStyle::parse("Fond blanc"), None);
    }
}
