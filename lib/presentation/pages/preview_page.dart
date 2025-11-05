import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:screenshot/screenshot.dart';
import '../../core/utils/save_to_gallery.dart';
import '../viewmodels/price_tag_viewmodel.dart';
import '../widgets/price_tag_widget.dart';

class PreviewPage extends ConsumerWidget {
  const PreviewPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tag = ref.watch(priceTagProvider);
    final controller = ScreenshotController();

    Future<void> saveImage(BuildContext context) async {
      try {
        final Uint8List? image = await controller.capture();
        if (image == null) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text("❌ Impossible de capturer l’image")),
          );
          return;
        }

        final success = await SaveToGallery.save(image);

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(success
                ? "✅ Image enregistrée dans la galerie"
                : "❌ Erreur lors de l’enregistrement"),
            behavior: SnackBarBehavior.floating,
          ),
        );
      } catch (e) {
        debugPrint("Erreur de capture : $e");
      }
    }

    return Scaffold(
      appBar: AppBar(title: const Text("Aperçu de l’étiquette")),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            Expanded(
              child: Center(
                child: Screenshot(
                  controller: controller,
                  child: PriceTagWidget(tag: tag),
                ),
              ),
            ),
            const SizedBox(height: 20),
            ElevatedButton.icon(
              icon: const Icon(Icons.download),
              label: const Text("Enregistrer dans la galerie"),
              onPressed: () => saveImage(context),
              style: ElevatedButton.styleFrom(
                minimumSize: const Size(double.infinity, 55),
                backgroundColor: Theme.of(context).colorScheme.primary,
                foregroundColor: Colors.white,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
