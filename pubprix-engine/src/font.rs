//! Police vectorielle minimale pour l'overlay prix/nom — RS-04.
//!
//! Aucune police (.ttf/.otf) n'est embarquée dans ce dépôt et on ne veut
//! pas dépendre d'une police système (portabilité en conteneur). On
//! dessine donc chaque caractère comme une poignée de segments (technique
//! "stick font", façon afficheur à segments) sur une grille de points
//! normalisée [0,1]x[0,1], rendus ensuite par tiny-skia. Couverture :
//! chiffres, A-Z, espace et une poignée de signes de ponctuation.
//! Volontairement simple, comme le reste de ce sprint.

pub type Point = (f32, f32);

pub enum Glyph {
    /// Une ou plusieurs polylignes à tracer.
    Strokes(&'static [&'static [Point]]),
    /// Point isolé (le "." par exemple).
    Dot,
    /// Rien à dessiner (espace).
    Empty,
}

const TL: Point = (0.0, 0.0);
const TM: Point = (0.5, 0.0);
const TR: Point = (1.0, 0.0);
const ML: Point = (0.0, 0.5);
const MM: Point = (0.5, 0.5);
const MR: Point = (1.0, 0.5);
const BL: Point = (0.0, 1.0);
const BM: Point = (0.5, 1.0);
const BR: Point = (1.0, 1.0);

const DIGIT_0: &[&[Point]] = &[&[TL, TR], &[TR, MR], &[MR, BR], &[BL, BR], &[ML, BL], &[TL, ML]];
const DIGIT_1: &[&[Point]] = &[&[TR, MR], &[MR, BR]];
const DIGIT_2: &[&[Point]] = &[&[TL, TR], &[TR, MR], &[ML, MR], &[ML, BL], &[BL, BR]];
const DIGIT_3: &[&[Point]] = &[&[TL, TR], &[TR, MR], &[ML, MR], &[MR, BR], &[BL, BR]];
const DIGIT_4: &[&[Point]] = &[&[TL, ML], &[ML, MR], &[TR, MR], &[MR, BR]];
const DIGIT_5: &[&[Point]] = &[&[TL, TR], &[TL, ML], &[ML, MR], &[MR, BR], &[BL, BR]];
const DIGIT_6: &[&[Point]] = &[&[TL, TR], &[TL, ML], &[ML, MR], &[ML, BL], &[MR, BR], &[BL, BR]];
const DIGIT_7: &[&[Point]] = &[&[TL, TR], &[TR, MR], &[MR, BR]];
const DIGIT_8: &[&[Point]] = &[&[TL, TR], &[TR, MR], &[MR, BR], &[BL, BR], &[ML, BL], &[TL, ML], &[ML, MR]];
const DIGIT_9: &[&[Point]] = &[&[TL, TR], &[TR, MR], &[MR, BR], &[BL, BR], &[TL, ML], &[ML, MR]];

const LETTER_A: &[&[Point]] = &[&[BL, TM, BR], &[ML, MR]];
const LETTER_B: &[&[Point]] = &[&[TL, BL], &[TL, TR, MR, ML], &[ML, MR, BR, BL]];
const LETTER_C: &[&[Point]] = &[&[TR, TL, BL, BR]];
const LETTER_D: &[&[Point]] = &[&[TL, BL], &[TL, TR, BR, BL]];
const LETTER_E: &[&[Point]] = &[&[TR, TL, BL, BR], &[ML, MR]];
const LETTER_F: &[&[Point]] = &[&[BL, TL, TR], &[ML, MR]];
const LETTER_G: &[&[Point]] = &[&[TR, TL, BL, BR, MR, MM]];
const LETTER_H: &[&[Point]] = &[&[TL, BL], &[TR, BR], &[ML, MR]];
const LETTER_I: &[&[Point]] = &[&[TL, TR], &[BL, BR], &[TM, BM]];
const LETTER_J: &[&[Point]] = &[&[TL, TR], &[TR, BR, BL]];
const LETTER_K: &[&[Point]] = &[&[TL, BL], &[ML, TR], &[ML, BR]];
const LETTER_L: &[&[Point]] = &[&[TL, BL, BR]];
const LETTER_M: &[&[Point]] = &[&[BL, TL, MM, TR, BR]];
const LETTER_N: &[&[Point]] = &[&[BL, TL, BR, TR]];
const LETTER_O: &[&[Point]] = &[&[TL, TR, BR, BL, TL]];
const LETTER_P: &[&[Point]] = &[&[BL, TL, TR, MR, ML]];
const LETTER_Q: &[&[Point]] = &[&[TL, TR, BR, BL, TL], &[MM, BR]];
const LETTER_R: &[&[Point]] = &[&[BL, TL, TR, MR, ML], &[ML, BR]];
const LETTER_S: &[&[Point]] = &[&[TR, TL, ML, MR, BR, BL]];
const LETTER_T: &[&[Point]] = &[&[TL, TR], &[TM, BM]];
const LETTER_U: &[&[Point]] = &[&[TL, BL, BR, TR]];
const LETTER_V: &[&[Point]] = &[&[TL, BM, TR]];
const LETTER_W: &[&[Point]] = &[&[TL, BL, MM, BR, TR]];
const LETTER_X: &[&[Point]] = &[&[TL, BR], &[TR, BL]];
const LETTER_Y: &[&[Point]] = &[&[TL, MM], &[TR, MM], &[MM, BM]];
const LETTER_Z: &[&[Point]] = &[&[TL, TR, BL, BR]];

const HYPHEN: &[&[Point]] = &[&[(0.15, 0.5), (0.85, 0.5)]];
const SLASH: &[&[Point]] = &[&[BL, TR]];
const COMMA: &[&[Point]] = &[&[BM, (0.65, 1.2)]];
const APOSTROPHE: &[&[Point]] = &[&[TM, (0.5, 0.35)]];
/// Repère de substitution pour un caractère non pris en charge (jamais de
/// panique : on affiche toujours quelque chose).
const FALLBACK: &[&[Point]] = &[&[TM, MR, BM, ML, TM]];

/// Normalise un caractère avant lookup : mise en capitale et
/// simplification des accents français courants.
pub fn normalize_char(c: char) -> char {
    let upper = c.to_uppercase().next().unwrap_or(c);
    match upper {
        'À' | 'Â' | 'Ä' => 'A',
        'É' | 'È' | 'Ê' | 'Ë' => 'E',
        'Î' | 'Ï' => 'I',
        'Ô' | 'Ö' => 'O',
        'Ù' | 'Û' | 'Ü' => 'U',
        'Ç' => 'C',
        other => other,
    }
}

pub fn glyph_for(c: char) -> Glyph {
    match normalize_char(c) {
        ' ' => Glyph::Empty,
        '.' => Glyph::Dot,
        ',' => Glyph::Strokes(COMMA),
        '-' => Glyph::Strokes(HYPHEN),
        '/' => Glyph::Strokes(SLASH),
        '\'' => Glyph::Strokes(APOSTROPHE),
        '0' => Glyph::Strokes(DIGIT_0),
        '1' => Glyph::Strokes(DIGIT_1),
        '2' => Glyph::Strokes(DIGIT_2),
        '3' => Glyph::Strokes(DIGIT_3),
        '4' => Glyph::Strokes(DIGIT_4),
        '5' => Glyph::Strokes(DIGIT_5),
        '6' => Glyph::Strokes(DIGIT_6),
        '7' => Glyph::Strokes(DIGIT_7),
        '8' => Glyph::Strokes(DIGIT_8),
        '9' => Glyph::Strokes(DIGIT_9),
        'A' => Glyph::Strokes(LETTER_A),
        'B' => Glyph::Strokes(LETTER_B),
        'C' => Glyph::Strokes(LETTER_C),
        'D' => Glyph::Strokes(LETTER_D),
        'E' => Glyph::Strokes(LETTER_E),
        'F' => Glyph::Strokes(LETTER_F),
        'G' => Glyph::Strokes(LETTER_G),
        'H' => Glyph::Strokes(LETTER_H),
        'I' => Glyph::Strokes(LETTER_I),
        'J' => Glyph::Strokes(LETTER_J),
        'K' => Glyph::Strokes(LETTER_K),
        'L' => Glyph::Strokes(LETTER_L),
        'M' => Glyph::Strokes(LETTER_M),
        'N' => Glyph::Strokes(LETTER_N),
        'O' => Glyph::Strokes(LETTER_O),
        'P' => Glyph::Strokes(LETTER_P),
        'Q' => Glyph::Strokes(LETTER_Q),
        'R' => Glyph::Strokes(LETTER_R),
        'S' => Glyph::Strokes(LETTER_S),
        'T' => Glyph::Strokes(LETTER_T),
        'U' => Glyph::Strokes(LETTER_U),
        'V' => Glyph::Strokes(LETTER_V),
        'W' => Glyph::Strokes(LETTER_W),
        'X' => Glyph::Strokes(LETTER_X),
        'Y' => Glyph::Strokes(LETTER_Y),
        'Z' => Glyph::Strokes(LETTER_Z),
        _ => Glyph::Strokes(FALLBACK),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn digits_and_letters_all_render_something() {
        for c in "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ".chars() {
            match glyph_for(c) {
                Glyph::Strokes(strokes) => assert!(!strokes.is_empty(), "glyphe vide pour '{c}'"),
                other => panic!("'{c}' devrait être un tracé, pas {other:?}", other = debug_kind(&other)),
            }
        }
    }

    #[test]
    fn space_is_empty_and_period_is_a_dot() {
        assert!(matches!(glyph_for(' '), Glyph::Empty));
        assert!(matches!(glyph_for('.'), Glyph::Dot));
    }

    #[test]
    fn unsupported_characters_fall_back_without_panicking() {
        for c in ['€', '#', '😀'] {
            match glyph_for(c) {
                Glyph::Strokes(strokes) => assert!(!strokes.is_empty()),
                other => panic!("caractère de secours inattendu: {other:?}", other = debug_kind(&other)),
            }
        }
    }

    #[test]
    fn accented_letters_normalize_to_their_base_letter() {
        assert_eq!(normalize_char('é'), 'E');
        assert_eq!(normalize_char('Ç'), 'C');
        assert_eq!(normalize_char('a'), 'A');
    }

    #[test]
    fn all_stroke_points_stay_within_the_unit_glyph_box() {
        for c in "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-/,'".chars() {
            if let Glyph::Strokes(strokes) = glyph_for(c) {
                for stroke in strokes {
                    for (x, y) in *stroke {
                        assert!((-0.01..=1.3).contains(x), "x hors bornes pour '{c}': {x}");
                        assert!((-0.01..=1.3).contains(y), "y hors bornes pour '{c}': {y}");
                    }
                }
            }
        }
    }

    fn debug_kind(g: &Glyph) -> &'static str {
        match g {
            Glyph::Strokes(_) => "Strokes",
            Glyph::Dot => "Dot",
            Glyph::Empty => "Empty",
        }
    }
}
