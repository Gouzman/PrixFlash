/// Miroir Dart de l'entité `product_draft` renvoyée par l'API
/// (voir pubprix-api, POST /drafts et PATCH /drafts/{id}).
class Draft {
  const Draft({
    required this.id,
    required this.sessionToken,
    this.name,
    this.price,
    this.currency,
  });

  factory Draft.fromJson(Map<String, dynamic> json) {
    return Draft(
      id: json['id'] as String,
      sessionToken: json['session_token'] as String,
      name: json['name'] as String?,
      price: (json['price'] as num?)?.toDouble(),
      currency: json['currency'] as String?,
    );
  }

  final String id;
  final String sessionToken;
  final String? name;
  final double? price;
  final String? currency;
}
