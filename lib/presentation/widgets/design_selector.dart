import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../viewmodels/price_tag_viewmodel.dart';
import 'price_tag_widget.dart';

class DesignSelector extends ConsumerWidget {
  const DesignSelector({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tag = ref.watch(priceTagProvider);
    final viewModel = ref.read(priceTagProvider.notifier);

    // Liste des designs à afficher
    final designs = [
      {'label': 'Carré', 'type': 'square'},
      {'label': 'Suspendu', 'type': 'hanging'},
      {'label': 'Ruban', 'type': 'ribbon'},
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),
        Text(
          "🧩 Choisir un modèle d’étiquette",
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
        ),
        const SizedBox(height: 12),
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: designs.map((d) {
              final selected = tag.designType == d['type'];
              final miniTag = tag.copyWith(designType: d['type']!);

              return GestureDetector(
                onTap: () => viewModel.updateDesignType(d['type']!),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 250),
                  margin: const EdgeInsets.symmetric(horizontal: 8),
                  padding: const EdgeInsets.all(8),
                  width: 120,
                  height: 120,
                  decoration: BoxDecoration(
                    color: selected
                        ? Theme.of(context).colorScheme.primary.withValues(alpha: 0.1)
                        : Colors.grey[100],
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: selected
                          ? Theme.of(context).colorScheme.primary
                          : Colors.grey.shade300,
                      width: selected ? 2 : 1,
                    ),
                  ),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Transform.scale(
                        scale: 0.6,
                        child: PriceTagWidget(tag: miniTag),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        d['label']!,
                        style: TextStyle(
                          fontWeight: FontWeight.w600,
                          color: selected
                              ? Theme.of(context).colorScheme.primary
                              : Colors.black87,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }).toList(),
          ),
        ),
      ],
    );
  }
}
