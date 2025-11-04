import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../viewmodels/price_tag_viewmodel.dart';

class LogoSelector extends ConsumerWidget {
  const LogoSelector({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final viewModel = ref.read(priceTagProvider.notifier);
    final tag = ref.watch(priceTagProvider);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text("📷 Logo (optionnel)"),
        const SizedBox(height: 8),
        Row(
          children: [
            ElevatedButton.icon(
              onPressed: () {
                // TODO: Implémenter la sélection d'image
                _showLogoOptions(context, viewModel);
              },
              icon: const Icon(Icons.add_photo_alternate),
              label: const Text('Ajouter logo'),
            ),
            if (tag.logoPath != null) ...[
              const SizedBox(width: 12),
              ElevatedButton.icon(
                onPressed: () => viewModel.updateLogoPath(null),
                icon: const Icon(Icons.delete),
                label: const Text('Supprimer'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.red[100],
                  foregroundColor: Colors.red[700],
                ),
              ),
            ],
          ],
        ),
        if (tag.logoPath != null) ...[
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              border: Border.all(color: Colors.grey[300]!),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                const Icon(Icons.image, color: Colors.grey),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    tag.logoPath!.split('/').last,
                    style: const TextStyle(fontSize: 12),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }

  void _showLogoOptions(BuildContext context, viewModel) {
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.camera_alt),
              title: const Text('Prendre une photo'),
              onTap: () {
                Navigator.pop(context);
                // TODO: Implémenter camera
                _simulateImageSelection(viewModel, 'camera_photo.jpg');
              },
            ),
            ListTile(
              leading: const Icon(Icons.photo_library),
              title: const Text('Choisir depuis la galerie'),
              onTap: () {
                Navigator.pop(context);
                // TODO: Implémenter galerie
                _simulateImageSelection(viewModel, 'gallery_image.jpg');
              },
            ),
          ],
        ),
      ),
    );
  }

  void _simulateImageSelection(viewModel, String filename) {
    // Simulation temporaire
    viewModel.updateLogoPath('assets/logos/$filename');
  }
}