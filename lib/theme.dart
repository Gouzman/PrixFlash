import 'package:flutter/material.dart';

/// Palette Pubprix — reprend les tokens du site vitrine (voir le cahier
/// des charges technique, section "Stack technique").
class PubprixColors {
  static const accent = Color(0xFFE2661B);

  static const darkBg = Color(0xFF0B0A08);
  static const darkSurface = Color(0xFF17140F);
  static const darkBorder = Color(0x17FFFFFF);
  static const darkText = Color(0xFFF2EBDE);
  static const darkTextMuted = Color(0xFFABA093);

  static const lightBg = Color(0xFFFBF6EF);
  static const lightSurface = Color(0xFFFFFFFF);
  static const lightBorder = Color(0x1A211C16);
  static const lightText = Color(0xFF211C16);
  static const lightTextMuted = Color(0xFF63564A);
}

ThemeData pubprixLightTheme() {
  return ThemeData(
    useMaterial3: true,
    brightness: Brightness.light,
    scaffoldBackgroundColor: PubprixColors.lightBg,
    colorScheme: ColorScheme.fromSeed(
      seedColor: PubprixColors.accent,
      brightness: Brightness.light,
      primary: PubprixColors.accent,
      surface: PubprixColors.lightSurface,
    ),
    textTheme: const TextTheme().apply(
      bodyColor: PubprixColors.lightText,
      displayColor: PubprixColors.lightText,
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: PubprixColors.accent,
        foregroundColor: Colors.white,
        shape: const StadiumBorder(),
        padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 16),
      ),
    ),
  );
}

ThemeData pubprixDarkTheme() {
  return ThemeData(
    useMaterial3: true,
    brightness: Brightness.dark,
    scaffoldBackgroundColor: PubprixColors.darkBg,
    colorScheme: ColorScheme.fromSeed(
      seedColor: PubprixColors.accent,
      brightness: Brightness.dark,
      primary: PubprixColors.accent,
      surface: PubprixColors.darkSurface,
    ),
    textTheme: const TextTheme().apply(
      bodyColor: PubprixColors.darkText,
      displayColor: PubprixColors.darkText,
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: PubprixColors.accent,
        foregroundColor: Colors.white,
        shape: const StadiumBorder(),
        padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 16),
      ),
    ),
  );
}
