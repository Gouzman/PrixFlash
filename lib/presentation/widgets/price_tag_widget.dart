import 'package:flutter/material.dart';
import '../../domain/models/price_tag_model.dart';
import 'price_tag_square.dart';
import 'price_tag_hanging.dart';
import 'price_tag_ribbon.dart';

class PriceTagWidget extends StatelessWidget {
  final PriceTag tag;
  const PriceTagWidget({super.key, required this.tag});

  @override
  Widget build(BuildContext context) {
    // Switch pour sélectionner le bon composant selon designType
    switch (tag.designType) {
      case 'hanging':
        return PriceTagHanging(tag: tag);
      case 'ribbon':
        return PriceTagRibbon(tag: tag);
      case 'square':
      default:
        return PriceTagSquare(tag: tag);
    }
  }
}
