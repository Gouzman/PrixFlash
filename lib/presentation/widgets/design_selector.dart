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

    final designs = [
      {'label': 'Carré', 'type': 'square'},
      {'label': 'Suspendu', 'type': 'hanging'},
      {'label': 'Ruban', 'type': 'ribbon'},
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 20),
        Text(
          "🧩 Choisir un modèle d’étiquette",
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 140, // Limite fixe pour le contenu
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            separatorBuilder: (_, __) => const SizedBox(width: 12),
            itemCount: designs.length,
            itemBuilder: (context, index) {
              final d = designs[index];
              final selected = tag.designType == d['type'];
              final miniTag = tag.copyWith(designType: d['type']!);

              return GestureDetector(
                onTap: () => viewModel.updateDesignType(d['type']!),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 250),
                  width: 120,
                  padding: const EdgeInsets.all(8),
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
                      Expanded(
                        child: FittedBox(
                          fit: BoxFit.contain,
                          child: PriceTagWidget(tag: miniTag),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        d['label']!,
                        style: TextStyle(
                          fontWeight: FontWeight.w600,
                          fontSize: 13,
                          color: selected
                              ? Theme.of(context).colorScheme.primary
                              : Colors.black87,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}
