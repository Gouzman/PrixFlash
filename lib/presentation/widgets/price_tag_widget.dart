import 'package:flutter/material.dart';
import '../../domain/models/price_tag_model.dart';

class PriceTagWidget extends StatelessWidget {
  final PriceTag tag;
  const PriceTagWidget({super.key, required this.tag});

  @override
  Widget build(BuildContext context) {
    final bgColor = Color(
      int.parse(tag.backgroundColor.replaceAll('#', '0xff')),
    );
    final textColor = Color(int.parse(tag.textColor.replaceAll('#', '0xff')));

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.grey.shade400, blurRadius: 6)],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            tag.productName,
            style: TextStyle(
              color: textColor,
              fontWeight: FontWeight.bold,
              fontSize: 22,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            "${tag.oldPrice.toStringAsFixed(0)} ${tag.currency}",
            style: TextStyle(
              fontSize: 16,
              decoration: TextDecoration.lineThrough,
              color: textColor.withOpacity(0.6),
            ),
          ),
          Text(
            "${tag.newPrice.toStringAsFixed(0)} ${tag.currency}",
            style: TextStyle(
              fontSize: 26,
              fontWeight: FontWeight.bold,
              color: Colors.redAccent,
            ),
          ),
        ],
      ),
    );
  }
}
