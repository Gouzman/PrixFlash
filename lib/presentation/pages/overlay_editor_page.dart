import 'package:flutter/material.dart';
import '../../domain/models/price_tag_model.dart';
import '../widgets/price_tag_widget.dart';

class OverlayEditorPage extends StatelessWidget {
  final PriceTag tag;

  const OverlayEditorPage({super.key, required this.tag});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text("Placer sur image produit")),
      body: Column(
        children: [
          Expanded(
            child: Container(
              width: double.infinity,
              color: Colors.grey[100],
              child: Stack(
                children: [
                  // Zone pour l'image produit (à implémenter)
                  Center(
                    child: Container(
                      width: 300,
                      height: 300,
                      decoration: BoxDecoration(
                        color: Colors.grey[300],
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.grey),
                      ),
                      child: const Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.add_a_photo,
                              size: 50,
                              color: Colors.grey,
                            ),
                            SizedBox(height: 8),
                            Text("Appuyez pour ajouter une image produit"),
                          ],
                        ),
                      ),
                    ),
                  ),
                  // Étiquette de prix draggable
                  Positioned(
                    top: 50,
                    right: 50,
                    child: Draggable(
                      feedback: Material(
                        child: SizedBox(
                          width: 200,
                          child: PriceTagWidget(tag: tag),
                        ),
                      ),
                      childWhenDragging: Opacity(
                        opacity: 0.5,
                        child: SizedBox(
                          width: 200,
                          child: PriceTagWidget(tag: tag),
                        ),
                      ),
                      child: SizedBox(
                        width: 200,
                        child: PriceTagWidget(tag: tag),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    icon: const Icon(Icons.photo_library),
                    label: const Text("Choisir image"),
                    onPressed: () {
                      // Implémenter la sélection d'image
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text("Fonctionnalité à venir")),
                      );
                    },
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: ElevatedButton.icon(
                    icon: const Icon(Icons.save),
                    label: const Text("Enregistrer"),
                    onPressed: () {
                      // implementer l'enregistrement
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text("Fonctionnalité à venir")),
                      );
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Theme.of(context).primaryColor,
                      foregroundColor: Colors.white,
                    ),
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
