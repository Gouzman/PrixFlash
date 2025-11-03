class PriceTag {
  final String productName;
  final double oldPrice;
  final double newPrice;
  final String currency;
  final String backgroundColor;
  final String textColor;

  const PriceTag({
    required this.productName,
    required this.oldPrice,
    required this.newPrice,
    this.currency = 'FCFA',
    this.backgroundColor = '#FFFFFF',
    this.textColor = '#000000',
  });

  PriceTag copyWith({
    String? productName,
    double? oldPrice,
    double? newPrice,
    String? currency,
    String? backgroundColor,
    String? textColor,
  }) {
    return PriceTag(
      productName: productName ?? this.productName,
      oldPrice: oldPrice ?? this.oldPrice,
      newPrice: newPrice ?? this.newPrice,
      currency: currency ?? this.currency,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      textColor: textColor ?? this.textColor,
    );
  }

  Map<String, dynamic> toJson() => {
    'productName': productName,
    'oldPrice': oldPrice,
    'newPrice': newPrice,
    'currency': currency,
    'backgroundColor': backgroundColor,
    'textColor': textColor,
  };
}
