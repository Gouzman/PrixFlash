class PriceTag {
  final String productName;
  final double oldPrice;
  final double newPrice;
  final String currency;
  final String backgroundColor;
  final String textColor;
  final String designType;
  final String fontFamily;   // 🆕
  final String? logoPath;    // 🆕

  const PriceTag({
    required this.productName,
    required this.oldPrice,
    required this.newPrice,
    this.currency = 'FCFA',
    this.backgroundColor = '#000000',
    this.textColor = '#FFFFFF',
    this.designType = 'square',
    this.fontFamily = 'Poppins',
    this.logoPath,
  });

  PriceTag copyWith({
    String? productName,
    double? oldPrice,
    double? newPrice,
    String? currency,
    String? backgroundColor,
    String? textColor,
    String? designType,
    String? fontFamily,
    String? logoPath,
  }) {
    return PriceTag(
      productName: productName ?? this.productName,
      oldPrice: oldPrice ?? this.oldPrice,
      newPrice: newPrice ?? this.newPrice,
      currency: currency ?? this.currency,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      textColor: textColor ?? this.textColor,
      designType: designType ?? this.designType,
      fontFamily: fontFamily ?? this.fontFamily,
      logoPath: logoPath ?? this.logoPath,
    );
  }
}
