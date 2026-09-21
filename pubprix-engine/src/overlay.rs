//! Overlay prix + nom du produit — RS-04.
//!
//! Bandeau semi-transparent en bas de l'image, texte tracé avec la police
//! vectorielle de font.rs, rendu via tiny-skia (anti-aliasing).

use image::RgbaImage;
use tiny_skia::{Color, LineCap, LineJoin, Paint, PathBuilder, Pixmap, Rect, Stroke, Transform};

use crate::font::{glyph_for, Glyph};

const BAR_HEIGHT_RATIO: f32 = 0.22;
const BAR_COLOR: (u8, u8, u8, u8) = (10, 10, 10, 170);
const TEXT_COLOR: Color = Color::WHITE;

pub fn draw_price_and_name(canvas: &RgbaImage, price_label: &str, name: Option<&str>) -> RgbaImage {
    let (w, h) = canvas.dimensions();
    let Some(mut pixmap) = Pixmap::new(w, h) else {
        return canvas.clone();
    };
    // L'image composée est toujours totalement opaque (alpha=255), donc
    // premultiplied == straight : la copie d'octets est directe.
    pixmap.data_mut().copy_from_slice(canvas.as_raw());

    let bar_h = (h as f32 * BAR_HEIGHT_RATIO).round();
    let bar_y = h as f32 - bar_h;
    draw_bar(&mut pixmap, bar_y, w as f32, bar_h);

    if let Some(name) = name.filter(|n| !n.trim().is_empty()) {
        let name_h = bar_h * 0.22;
        let price_h = bar_h * 0.42;
        draw_text_line(&mut pixmap, price_label, w as f32 / 2.0, bar_y + bar_h * 0.14, price_h);
        draw_text_line(&mut pixmap, &name.to_uppercase(), w as f32 / 2.0, bar_y + bar_h * 0.68, name_h);
    } else {
        let price_h = bar_h * 0.5;
        draw_text_line(&mut pixmap, price_label, w as f32 / 2.0, bar_y + bar_h * 0.28, price_h);
    }

    RgbaImage::from_raw(w, h, pixmap.data().to_vec()).unwrap_or_else(|| canvas.clone())
}

fn draw_bar(pixmap: &mut Pixmap, y: f32, w: f32, h: f32) {
    let mut paint = Paint::default();
    paint.set_color_rgba8(BAR_COLOR.0, BAR_COLOR.1, BAR_COLOR.2, BAR_COLOR.3);
    paint.anti_alias = true;
    if let Some(rect) = Rect::from_xywh(0.0, y, w, h) {
        pixmap.fill_rect(rect, &paint, Transform::identity(), None);
    }
}

/// Dessine `text` centré horizontalement sur `center_x`, avec le haut de
/// chaque glyphe posé à `top_y`, chaque glyphe occupant une hauteur
/// `glyph_h` (largeur = 0.62 * hauteur, comme une police condensée).
fn draw_text_line(pixmap: &mut Pixmap, text: &str, center_x: f32, top_y: f32, glyph_h: f32) {
    if glyph_h <= 0.0 {
        return;
    }
    let glyph_w = glyph_h * 0.62;
    let gap = glyph_w * 0.35;
    let chars: Vec<char> = text.chars().collect();
    if chars.is_empty() {
        return;
    }

    let total_w = chars.len() as f32 * glyph_w + (chars.len() as f32 - 1.0).max(0.0) * gap;
    let mut cursor_x = center_x - total_w / 2.0;

    let mut paint = Paint::default();
    paint.set_color(TEXT_COLOR);
    paint.anti_alias = true;
    let stroke = Stroke {
        width: (glyph_w * 0.16).max(1.0),
        line_cap: LineCap::Round,
        line_join: LineJoin::Round,
        ..Default::default()
    };

    for c in chars {
        match glyph_for(c) {
            Glyph::Empty => {}
            Glyph::Dot => {
                if let Some(path) = PathBuilder::from_circle(
                    cursor_x + glyph_w / 2.0,
                    top_y + glyph_h * 0.92,
                    (stroke.width / 2.0).max(1.0),
                ) {
                    pixmap.fill_path(&path, &paint, tiny_skia::FillRule::Winding, Transform::identity(), None);
                }
            }
            Glyph::Strokes(strokes) => {
                for stroke_points in strokes {
                    let mut pb = PathBuilder::new();
                    let mut points = stroke_points.iter();
                    if let Some(&(x0, y0)) = points.next() {
                        pb.move_to(cursor_x + x0 * glyph_w, top_y + y0 * glyph_h);
                        for &(x, y) in points {
                            pb.line_to(cursor_x + x * glyph_w, top_y + y * glyph_h);
                        }
                    }
                    if let Some(path) = pb.finish() {
                        pixmap.stroke_path(&path, &paint, &stroke, Transform::identity(), None);
                    }
                }
            }
        }
        cursor_x += glyph_w + gap;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use image::Rgba;

    #[test]
    fn overlay_keeps_canvas_dimensions() {
        let canvas = RgbaImage::from_pixel(400, 400, Rgba([50, 50, 50, 255]));
        let result = draw_price_and_name(&canvas, "12 500 FCFA", Some("Sac a main"));
        assert_eq!(result.dimensions(), (400, 400));
    }

    #[test]
    fn overlay_darkens_the_bottom_bar_area() {
        let canvas = RgbaImage::from_pixel(200, 200, Rgba([255, 255, 255, 255]));
        let result = draw_price_and_name(&canvas, "1 000 FCFA", None);
        // Un pixel dans le bandeau, loin de tout trait de texte, doit
        // être assombri par l'overlay semi-transparent.
        let pixel = result.get_pixel(5, 199);
        assert!(pixel[0] < 200, "le bandeau ne semble pas avoir été assombri: {pixel:?}");
    }

    #[test]
    fn overlay_leaves_area_above_the_bar_untouched() {
        let canvas = RgbaImage::from_pixel(200, 200, Rgba([255, 255, 255, 255]));
        let result = draw_price_and_name(&canvas, "1 000 FCFA", None);
        assert_eq!(*result.get_pixel(5, 0), Rgba([255, 255, 255, 255]));
    }

    #[test]
    fn overlay_does_not_panic_on_empty_or_unsupported_text() {
        let canvas = RgbaImage::from_pixel(100, 100, Rgba([0, 0, 0, 255]));
        let _ = draw_price_and_name(&canvas, "", Some(""));
        let _ = draw_price_and_name(&canvas, "€€€ 😀", Some("???"));
    }
}
