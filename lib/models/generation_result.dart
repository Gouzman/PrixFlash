/// Miroir Dart de la réponse du moteur de génération
/// (voir pubprix-engine, POST /generate — status + image_url).
class GenerationResult {
  const GenerationResult({required this.status, this.imageUrl});

  factory GenerationResult.fromJson(Map<String, dynamic> json) {
    return GenerationResult(
      status: json['status'] as String,
      imageUrl: json['image_url'] as String?,
    );
  }

  final String status;
  final String? imageUrl;

  bool get isReady => status == 'ok' && imageUrl != null;
}
