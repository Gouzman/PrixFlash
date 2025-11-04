import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:screenshot/screenshot.dart';
import 'package:image_gallery_saver/image_gallery_saver.dart';
import '../viewmodels/price_tag_viewmodel.dart';
import '../widgets/price_tag_widget.dart';
import 'overlay_editor_page.dart';

class PreviewPage extends ConsumerWidget {
  const PreviewPage({super.key});

  Future<void> _saveToGallery(Uint8List image, BuildContext context) async {
    try {
      await ImageGallerySaver.saveImage(
        image,
        name: "PrixFlash_${DateTime.now().millisecondsSinceEpoch}",
      );
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text("✅ Image enregistrée dans la galerie")),
        );
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text("❌ Erreur lors de l'enregistrement")),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tag = ref.watch(priceTagProvider);
    final controller = ScreenshotController();

    return Scaffold(
      appBar: AppBar(title: const Text("Aperçu de l’étiquette")),
      body: Column(
        children: [
          Expanded(
            child: Center(
              child: Screenshot(
                controller: controller,
                child: PriceTagWidget(tag: tag),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                ElevatedButton.icon(
                  icon: const Icon(Icons.download),
                  label: const Text("Enregistrer dans la galerie"),
                  onPressed: () async {
                    final image = await controller.capture();
                    if (image != null && context.mounted) {
                      await _saveToGallery(image, context);
                    }
                  },
                  style: ElevatedButton.styleFrom(
                    minimumSize: const Size(double.infinity, 50),
                  ),
                ),
                const SizedBox(height: 12),
                ElevatedButton.icon(
                  icon: const Icon(Icons.layers),
                  label: const Text("Placer sur image produit"),
                  onPressed: () => Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => OverlayEditorPage(tag: tag),
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    minimumSize: const Size(double.infinity, 50),
                    backgroundColor: Colors.deepOrangeAccent,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
