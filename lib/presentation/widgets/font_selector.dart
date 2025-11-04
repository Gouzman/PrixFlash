import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../viewmodels/price_tag_viewmodel.dart';

class FontSelector extends ConsumerWidget {
  const FontSelector({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final viewModel = ref.read(priceTagProvider.notifier);
    final tag = ref.watch(priceTagProvider);

    final fonts = {
      "Modern": GoogleFonts.poppins().fontFamily!,
      "Bold": GoogleFonts.robotoCondensed().fontFamily!,
      "Elegant": GoogleFonts.playfairDisplay().fontFamily!,
      "Digital": GoogleFonts.orbitron().fontFamily!,
      "Classic": GoogleFonts.lato().fontFamily!,
    };

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),
        const Text("🖋️ Style d’écriture"),
        const SizedBox(height: 8),
        Wrap(
          spacing: 10,
          children: fonts.entries.map((entry) {
            final selected = tag.fontFamily == entry.value;
            return GestureDetector(
              onTap: () => viewModel.updateFont(entry.value),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: selected
                      ? Theme.of(context).colorScheme.primary.withValues(alpha: 0.1)
                      : Colors.grey[100],
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: selected
                        ? Theme.of(context).colorScheme.primary
                        : Colors.grey.shade400,
                  ),
                ),
                child: Text(
                  entry.key,
                  style: TextStyle(
                    fontFamily: entry.value,
                    fontWeight: FontWeight.w600,
                    color: selected
                        ? Theme.of(context).colorScheme.primary
                        : Colors.black87,
                  ),
                ),
              ),
            );
          }).toList(),
        ),
      ],
    );
  }
}
