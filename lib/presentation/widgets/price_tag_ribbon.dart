import 'package:flutter/material.dart';
import '../../domain/models/price_tag_model.dart';

class PriceTagRibbon extends StatelessWidget {
  final PriceTag tag;
  const PriceTagRibbon({super.key, required this.tag});

  @override
  Widget build(BuildContext context) {
    final bg = Color(int.parse(tag.backgroundColor.replaceAll('#', '0xff')));
    final txt = Color(int.parse(tag.textColor.replaceAll('#', '0xff')));

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(6),
        boxShadow: const [BoxShadow(color: Colors.black26, blurRadius: 3)],
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(tag.newPrice.toStringAsFixed(0),
              style: TextStyle(color: txt, fontSize: 26, fontWeight: FontWeight.bold)),
          Text(" ${tag.currency}",
              style: TextStyle(color: txt, fontSize: 14)),
          const SizedBox(width: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              "${_getDiscount(tag)}% OFF",
              style: TextStyle(
                color: txt,
                fontSize: 10,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _getDiscount(PriceTag tag) {
    if (tag.oldPrice <= 0) return '0';
    final reduction = ((1 - tag.newPrice / tag.oldPrice) * 100).clamp(0, 100);
    return reduction.toStringAsFixed(0);
  }
}
