import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../viewmodels/price_tag_viewmodel.dart';
import '../widgets/price_tag_widget.dart';
import '../widgets/font_selector.dart';
import '../widgets/logo_selector.dart';
import 'tag_preview_page.dart';

class EditorPage extends ConsumerWidget {
  const EditorPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tag = ref.watch(priceTagProvider);
    final viewModel = ref.read(priceTagProvider.notifier);

    final bgColors = ['#000000', '#ff9800', '#ffffff', '#4caf50', '#2196f3'];
    final textColors = ['#ffffff', '#000000', '#f44336', '#ffeb3b', '#00bcd4'];

        // Liste des designs
    final designs = [
      {'name': 'Carré', 'type': 'square'},
      {'name': 'Suspendu', 'type': 'hanging'},
      {'name': 'Ruban', 'type': 'ribbon'},
    ];

    return Scaffold(
      appBar: AppBar(title: const Text("Créer une étiquette")),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            PriceTagWidget(tag: tag),
            const SizedBox(height: 24),

            TextField(
              decoration: const InputDecoration(
                labelText: "Nom du produit",
                border: OutlineInputBorder(),
              ),
              onChanged: viewModel.updateProductName,
            ),
            const SizedBox(height: 16),

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
            const SizedBox(height: 24),

            // Sélecteur de design
            const Text("🧩 Modèle d'étiquette"),
            const SizedBox(height: 10),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: designs.map((d) {
                final selected = tag.designType == d['type'];
                return GestureDetector(
                  onTap: () => viewModel.updateDesignType(d['type']!),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 300),
                    margin: const EdgeInsets.symmetric(horizontal: 8),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    decoration: BoxDecoration(
                      color: selected ? Theme.of(context).colorScheme.primary : Colors.grey[200],
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      d['name']!,
                      style: TextStyle(
                        color: selected ? Colors.white : Colors.black87,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
            
            // Sélecteur de police
            const FontSelector(),
            const SizedBox(height: 24),

            // Sélecteur de logo
            const LogoSelector(),
            const SizedBox(height: 24),

            // Couleur fond
            const Text("🎨 Couleur de fond"),
            const SizedBox(height: 10),
            Wrap(
              spacing: 10,
              children: bgColors.map((hex) {
                final color = Color(int.parse(hex.replaceAll('#', '0xff')));
                return GestureDetector(
                  onTap: () => viewModel.updateBgColor(hex),
                  child: CircleAvatar(
                    backgroundColor: color,
                    radius: 18,
                    child: tag.backgroundColor == hex
                        ? const Icon(Icons.check, color: Colors.white)
                        : null,
                  ),
                );
              }).toList(),
            ),

            const SizedBox(height: 20),

            // Couleur texte
            const Text("🖋️ Couleur du texte"),
            const SizedBox(height: 10),
            Wrap(
              spacing: 10,
              children: textColors.map((hex) {
                final color = Color(int.parse(hex.replaceAll('#', '0xff')));
                return GestureDetector(
                  onTap: () => viewModel.updateTextColor(hex),
                  child: CircleAvatar(
                    backgroundColor: color,
                    radius: 18,
                    child: tag.textColor == hex
                        ? const Icon(Icons.check, color: Colors.black)
                        : null,
                  ),
                );
              }).toList(),
            ),
            
            const SizedBox(height: 32),
            
            // Boutons d'action
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () => _saveTag(context, ref),
                    icon: const Icon(Icons.save),
                    label: const Text('Sauvegarder'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Theme.of(context).colorScheme.primary,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () => _exportTag(context, ref),
                    icon: const Icon(Icons.share),
                    label: const Text('Exporter'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Theme.of(context).colorScheme.secondary,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                  ),
                ),
              ],
            ),
            
            const SizedBox(height: 16),
            
            ElevatedButton.icon(
              onPressed: () => _previewTag(context, ref),
              icon: const Icon(Icons.visibility),
              label: const Text('Aperçu plein écran'),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.grey[100],
                foregroundColor: Colors.black87,
                padding: const EdgeInsets.symmetric(vertical: 12),
                minimumSize: const Size(double.infinity, 48),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _saveTag(BuildContext context, WidgetRef ref) {
    final tag = ref.read(priceTagProvider);
    // TODO: Implémenter la sauvegarde
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Étiquette "${tag.productName}" sauvegardée !'),
        backgroundColor: Colors.green,
      ),
    );
  }

  void _exportTag(BuildContext context, WidgetRef ref) {
    final tag = ref.read(priceTagProvider);
    // TODO: Implémenter l'export (PDF, image, etc.)
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Export de "${tag.productName}" en cours...'),
        backgroundColor: Colors.blue,
      ),
    );
  }

  void _previewTag(BuildContext context, WidgetRef ref) {
    final tag = ref.read(priceTagProvider);
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => TagPreviewPage(tag: tag),
      ),
    );
  }
}
