//! Mise en page de la publication — RS-03.
//!
//! Un seul template pour ce sprint (volontairement simple, pas de choix
//! de style) : 4 photos ou plus -> une photo principale + 3 vignettes ;
//! moins de 4 -> grille simple qui s'adapte au nombre de photos. Les
//! photos au-delà de la 4e ne sont pas utilisées par ce template pour
//! l'instant.
//!
//! Fonctions pures (aucun I/O) pour rester facilement testables.

use image::{imageops, DynamicImage, GenericImageView, Rgba, RgbaImage};

#[derive(Clone, Copy)]
struct Cell {
    x: u32,
    y: u32,
    w: u32,
    h: u32,
}

/// Compose les photos sur un canevas `canvas_w x canvas_h` selon le
/// nombre de photos disponibles.
pub fn compose_canvas(photos: &[DynamicImage], canvas_w: u32, canvas_h: u32) -> RgbaImage {
    let mut canvas = RgbaImage::from_pixel(canvas_w, canvas_h, Rgba([0, 0, 0, 255]));
    let cells = layout_cells(photos.len(), canvas_w, canvas_h);

    for (photo, cell) in photos.iter().zip(cells.iter()) {
        let fitted = cover_fit(photo, cell.w, cell.h);
        imageops::overlay(&mut canvas, &fitted, cell.x as i64, cell.y as i64);
    }

    canvas
}

/// Calcule les zones (cellules) où placer chaque photo, au plus 4 cellules.
fn layout_cells(photo_count: usize, w: u32, h: u32) -> Vec<Cell> {
    match photo_count.min(4) {
        0 => vec![],
        1 => vec![Cell { x: 0, y: 0, w, h }],
        2 => {
            let half_h = h / 2;
            vec![
                Cell { x: 0, y: 0, w, h: half_h },
                Cell { x: 0, y: half_h, w, h: h - half_h },
            ]
        }
        3 => {
            let top_h = h / 2;
            let bottom_h = h - top_h;
            let half_w = w / 2;
            vec![
                Cell { x: 0, y: 0, w, h: top_h },
                Cell { x: 0, y: top_h, w: half_w, h: bottom_h },
                Cell { x: half_w, y: top_h, w: w - half_w, h: bottom_h },
            ]
        }
        _ => {
            // Photo principale à gauche (65%), 3 vignettes empilées à droite.
            let main_w = (w as f32 * 0.65) as u32;
            let thumb_w = w - main_w;
            let thumb_h = h / 3;
            vec![
                Cell { x: 0, y: 0, w: main_w, h },
                Cell { x: main_w, y: 0, w: thumb_w, h: thumb_h },
                Cell { x: main_w, y: thumb_h, w: thumb_w, h: thumb_h },
                Cell { x: main_w, y: thumb_h * 2, w: thumb_w, h: h - thumb_h * 2 },
            ]
        }
    }
}

/// Redimensionne puis recadre l'image au centre pour couvrir exactement
/// `target_w x target_h` (comportement "object-fit: cover"), sans jamais
/// modifier l'image source passée en paramètre.
fn cover_fit(photo: &DynamicImage, target_w: u32, target_h: u32) -> RgbaImage {
    if target_w == 0 || target_h == 0 {
        return RgbaImage::new(0, 0);
    }

    let (src_w, src_h) = photo.dimensions();
    let scale = (target_w as f32 / src_w as f32).max(target_h as f32 / src_h as f32);
    let resized_w = (src_w as f32 * scale).ceil().max(1.0) as u32;
    let resized_h = (src_h as f32 * scale).ceil().max(1.0) as u32;

    let resized = photo.resize_exact(resized_w, resized_h, imageops::FilterType::Lanczos3);

    let crop_x = (resized_w.saturating_sub(target_w)) / 2;
    let crop_y = (resized_h.saturating_sub(target_h)) / 2;

    imageops::crop_imm(&resized, crop_x, crop_y, target_w, target_h).to_image()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn solid(color: [u8; 4], w: u32, h: u32) -> DynamicImage {
        DynamicImage::ImageRgba8(RgbaImage::from_pixel(w, h, Rgba(color)))
    }

    #[test]
    fn compose_canvas_has_requested_dimensions() {
        let photos = vec![solid([255, 0, 0, 255], 200, 200)];
        let canvas = compose_canvas(&photos, 1080, 1080);
        assert_eq!(canvas.dimensions(), (1080, 1080));
    }

    #[test]
    fn single_photo_fills_whole_canvas() {
        let photos = vec![solid([10, 20, 30, 255], 50, 50)];
        let canvas = compose_canvas(&photos, 100, 100);
        assert_eq!(*canvas.get_pixel(0, 0), Rgba([10, 20, 30, 255]));
        assert_eq!(*canvas.get_pixel(99, 99), Rgba([10, 20, 30, 255]));
    }

    #[test]
    fn two_photos_split_top_and_bottom() {
        let photos = vec![solid([255, 0, 0, 255], 40, 40), solid([0, 0, 255, 255], 40, 40)];
        let canvas = compose_canvas(&photos, 100, 100);
        assert_eq!(*canvas.get_pixel(50, 10), Rgba([255, 0, 0, 255]));
        assert_eq!(*canvas.get_pixel(50, 90), Rgba([0, 0, 255, 255]));
    }

    #[test]
    fn four_photos_use_main_plus_three_thumbnails_layout() {
        let cells = layout_cells(4, 1000, 900);
        assert_eq!(cells.len(), 4);
        assert_eq!(cells[0].x, 0);
        assert!(cells[0].w > cells[1].w, "la photo principale doit être plus large que les vignettes");
        assert_eq!(cells[1].x, cells[2].x);
        assert_eq!(cells[2].x, cells[3].x);
    }

    #[test]
    fn extra_photos_beyond_four_are_ignored_by_the_layout() {
        assert_eq!(layout_cells(6, 1000, 900).len(), 4);
    }

    #[test]
    fn cover_fit_always_returns_exact_target_size() {
        let photo = solid([1, 2, 3, 255], 300, 100);
        let fitted = cover_fit(&photo, 120, 200);
        assert_eq!(fitted.dimensions(), (120, 200));
    }

    #[test]
    fn cover_fit_does_not_mutate_the_source_image() {
        let photo = solid([9, 9, 9, 255], 64, 64);
        let before = photo.clone();
        let _ = cover_fit(&photo, 32, 96);
        assert_eq!(photo.as_bytes(), before.as_bytes());
    }
}
