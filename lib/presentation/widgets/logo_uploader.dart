import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../viewmodels/price_tag_viewmodel.dart';

class LogoUploader extends ConsumerWidget {
  const LogoUploader({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tag = ref.watch(priceTagProvider);
    final viewModel = ref.read(priceTagProvider.notifier);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),
        const Text(
          "🖼️ Logo du commerçant",
          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 16),
        
        Row(
          children: [
            // Effet visuel pour indiquer qu'il y a un logo chargé
            Container(
              decoration: BoxDecoration(
                border: Border.all(color: Colors.grey.shade300),
                borderRadius: BorderRadius.circular(8),
              ),
              child: tag.logoPath != null && File(tag.logoPath!).existsSync()
                  ? Image.file(File(tag.logoPath!), width: 60, height: 60, fit: BoxFit.cover)
                  : const Icon(Icons.image_outlined, color: Colors.grey, size: 40),
            ),
            
            const SizedBox(width: 16),
            
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (tag.logoPath != null)
                    Text(
                      'Logo ajouté ✓',
                      style: TextStyle(
                        color: Colors.green[700],
                        fontWeight: FontWeight.w500,
                      ),
                    )
                  else
                    const Text(
                      'Aucun logo sélectionné',
                      style: TextStyle(color: Colors.grey),
                    ),
                  
                  const SizedBox(height: 8),
                  
                  Row(
                    children: [
                      ElevatedButton.icon(
                        onPressed: () => _pickImage(context, viewModel),
                        icon: const Icon(Icons.add_photo_alternate, size: 18),
                        label: const Text('Ajouter'),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Theme.of(context).colorScheme.primary,
                          foregroundColor: Colors.white,
                          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                        ),
                      ),
                      
                      if (tag.logoPath != null) ...[
                        const SizedBox(width: 8),
                        ElevatedButton.icon(
                          onPressed: () => viewModel.updateLogoPath(null),
                          icon: const Icon(Icons.delete, size: 18),
                          label: const Text('Supprimer'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.red[100],
                            foregroundColor: Colors.red[700],
                            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                          ),
                        ),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
        
        if (tag.logoPath != null && File(tag.logoPath!).existsSync()) ...[
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.green[50],
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.green[200]!),
            ),
            child: Row(
              children: [
                Icon(Icons.check_circle, color: Colors.green[700], size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Logo prêt à être intégré dans l\'étiquette',
                    style: TextStyle(
                      color: Colors.green[700],
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }

  void _pickImage(BuildContext context, viewModel) async {
    final picker = ImagePicker();
    
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text(
              'Choisir une source',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 20),
            
            ListTile(
              leading: const Icon(Icons.camera_alt, color: Colors.blue),
              title: const Text('Appareil photo'),
              subtitle: const Text('Prendre une nouvelle photo'),
              onTap: () async {
                Navigator.pop(context);
                final XFile? image = await picker.pickImage(source: ImageSource.camera);
                if (image != null) {
                  viewModel.updateLogoPath(image.path);
                }
              },
            ),
            
            ListTile(
              leading: const Icon(Icons.photo_library, color: Colors.green),
              title: const Text('Galerie'),
              subtitle: const Text('Choisir depuis la galerie'),
              onTap: () async {
                Navigator.pop(context);
                final XFile? image = await picker.pickImage(source: ImageSource.gallery);
                if (image != null) {
                  viewModel.updateLogoPath(image.path);
                }
              },
            ),
          ],
        ),
      ),
    );
  }
}
