import 'package:flutter/material.dart';
import '../../domain/models/price_tag_model.dart';

class PriceTagHanging extends StatelessWidget {
  final PriceTag tag;
  const PriceTagHanging({super.key, required this.tag});

  @override
  Widget build(BuildContext context) {
    final bg = Color(int.parse(tag.backgroundColor.replaceAll('#', '0xff')));
    final txt = Color(int.parse(tag.textColor.replaceAll('#', '0xff')));

    return Stack(
      alignment: Alignment.topCenter,
      clipBehavior: Clip.none,
      children: [
        Container(
          width: 160,
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: bg,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            children: [
              Text(tag.productName,
                  style: TextStyle(color: txt, fontWeight: FontWeight.bold)),
              const SizedBox(height: 4),
              Text("${tag.newPrice.toStringAsFixed(0)} ${tag.currency}",
                  style: TextStyle(
                      color: txt, fontWeight: FontWeight.bold, fontSize: 22)),
              Text("${tag.oldPrice.toStringAsFixed(0)} ${tag.currency}",
                  style: TextStyle(
                      color: txt.withValues(alpha: 0.5),
                      decoration: TextDecoration.lineThrough)),
            ],
          ),
        ),
        Positioned(
          top: -10,
          child: Container(
            width: 20,
            height: 20,
            decoration: BoxDecoration(
              color: Colors.grey[300],
              shape: BoxShape.circle,
              border: Border.all(color: Colors.black12),
            ),
          ),
        ),
      ],
    );
  }
}
