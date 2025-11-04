import 'package:flutter/material.dart';
import '../../domain/models/price_tag_model.dart';

class PriceTagSquare extends StatelessWidget {
  final PriceTag tag;
  const PriceTagSquare({super.key, required this.tag});

  // Fonction pour parser les couleurs en toute sécurité
  Color parseColor(String hexColor, Color defaultColor) {
    try {
      String cleanHex = hexColor.replaceAll('#', '');
      if (cleanHex.length == 6 &&
          RegExp(r'^[0-9A-Fa-f]+$').hasMatch(cleanHex)) {
        return Color(int.parse('0xff$cleanHex'));
      }
      return defaultColor;
    } catch (e) {
      return defaultColor;
    }
  }

  @override
  Widget build(BuildContext context) {
    final bgColor = parseColor(tag.backgroundColor, Colors.white);
    final textColor = parseColor(tag.textColor, Colors.black);

    return Stack(
      clipBehavior: Clip.none,
      alignment: Alignment.topRight,
      children: [
        Container(
          width: 200,
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: bgColor,
            borderRadius: BorderRadius.circular(12),
            boxShadow: const [
              BoxShadow(
                color: Colors.black26,
                blurRadius: 6,
                offset: Offset(2, 3),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                tag.productName,
                textAlign: TextAlign.center,
                style: TextStyle(
                  color: textColor,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                  fontFamily: tag.fontFamily,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                "${tag.oldPrice.toStringAsFixed(2)} ${tag.currency}",
                style: TextStyle(
                  color: textColor.withValues(alpha: 0.6),
                  decoration: TextDecoration.lineThrough,
                  fontSize: 14,
                ),
              ),
              Text(
                "${tag.newPrice.toStringAsFixed(2)} ${tag.currency}",
                style: TextStyle(
                  color: textColor,
                  fontSize: 28,
                  fontWeight: FontWeight.bold,
                  fontFamily: tag.fontFamily,
                ),
              ),
              const SizedBox(height: 6),
              Container(
                padding: const EdgeInsets.symmetric(
                  vertical: 4,
                  horizontal: 10,
                ),
                decoration: BoxDecoration(
                  color: Colors.orangeAccent,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Text(
                  "LIMITED OFFER",
                  style: TextStyle(
                    color: Colors.black,
                    fontWeight: FontWeight.bold,
                    fontSize: 10,
                    letterSpacing: 1.1,
                  ),
                ),
              ),
            ],
          ),
        ),
        Positioned(
          top: -14,
          right: -10,
          child: Container(
            padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 12),
            decoration: BoxDecoration(
              color: Colors.orangeAccent,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(
              _getDiscountText(tag),
              style: const TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.bold,
                fontSize: 12,
              ),
            ),
          ),
        ),
      ],
    );
  }

  String _getDiscountText(PriceTag tag) {
    if (tag.oldPrice <= 0) return "NEW";
    final reduction = ((1 - tag.newPrice / tag.oldPrice) * 100).clamp(0, 100);
    return "-${reduction.toStringAsFixed(0)}%";
  }
}
