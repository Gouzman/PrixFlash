import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../viewmodels/price_tag_viewmodel.dart';
import '../widgets/price_tag_widget.dart';

class EditorPage extends ConsumerWidget {
  const EditorPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tag = ref.watch(priceTagProvider);
    final viewModel = ref.read(priceTagProvider.notifier);

    return Scaffold(
      appBar: AppBar(title: const Text("Créer une étiquette")),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text("Aperçu", style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            PriceTagWidget(tag: tag),
            const SizedBox(height: 24),

            // Champ nom produit
            TextField(
              decoration: const InputDecoration(
                labelText: "Nom du produit",
                border: OutlineInputBorder(),
              ),
              onChanged: viewModel.updateProductName,
            ),
            const SizedBox(height: 16),

            // Prix avant / après
            Row(
              children: [
                Expanded(
                  child: TextField(
                    decoration: const InputDecoration(
                      labelText: "Ancien prix",
                      border: OutlineInputBorder(),
                    ),
                    keyboardType: TextInputType.number,
                    onChanged: viewModel.updateOldPrice,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: TextField(
                    decoration: const InputDecoration(
                      labelText: "Nouveau prix",
                      border: OutlineInputBorder(),
                    ),
                    keyboardType: TextInputType.number,
                    onChanged: viewModel.updateNewPrice,
                  ),
                ),
              ],
            ),

            const SizedBox(height: 16),

            // Couleurs
            Row(
              children: [
                Expanded(
                  child: TextField(
                    decoration: const InputDecoration(
                      labelText: "Couleur de fond (#HEX)",
                      border: OutlineInputBorder(),
                    ),
                    onChanged: viewModel.updateBgColor,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: TextField(
                    decoration: const InputDecoration(
                      labelText: "Couleur du texte (#HEX)",
                      border: OutlineInputBorder(),
                    ),
                    onChanged: viewModel.updateTextColor,
                  ),
                ),
              ],
            ),

            const SizedBox(height: 30),

            // Bouton de génération
            Center(
              child: ElevatedButton.icon(
                icon: const Icon(Icons.preview),
                label: const Text("Générer l’aperçu"),
                onPressed: () => context.push('/preview'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Theme.of(context).colorScheme.primary,
                  foregroundColor: Colors.white,
                  minimumSize: const Size(200, 50),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
