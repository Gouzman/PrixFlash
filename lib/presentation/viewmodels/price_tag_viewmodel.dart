import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/models/price_tag_model.dart';

final priceTagProvider = NotifierProvider<PriceTagViewModel, PriceTag>(() {
  return PriceTagViewModel();
});

class PriceTagViewModel extends Notifier<PriceTag> {
  @override
  PriceTag build() {
    return const PriceTag(productName: '', oldPrice: 0, newPrice: 0);
  }

  void updateProductName(String name) {
    state = state.copyWith(productName: name);
  }

  void updateOldPrice(String price) {
    state = state.copyWith(oldPrice: double.tryParse(price) ?? 0);
  }

  void updateNewPrice(String price) {
    state = state.copyWith(newPrice: double.tryParse(price) ?? 0);
  }

  void updateBgColor(String color) {
    state = state.copyWith(backgroundColor: color);
  }

  void updateTextColor(String color) {
    state = state.copyWith(textColor: color);
  }
}
