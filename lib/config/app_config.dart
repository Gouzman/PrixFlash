/// Configuration d'environnement de l'app.
///
/// L'URL de base de l'API n'est jamais codée en dur : elle se règle via
/// `--dart-define=API_BASE_URL=https://...` au build/run. À défaut, on
/// retombe sur l'API locale utilisée en développement (voir pubprix-api).
class AppConfig {
  const AppConfig._();

  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );
}
