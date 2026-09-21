import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import 'price_screen.dart';

const int _minPhotos = 2;
const int _maxPhotos = 6;

/// Écran de sélection des photos — FL-02.
///
/// Sélection multiple via image_picker, aperçu en grille avec possibilité
/// de retirer une photo, et validation 2 à 6 photos avant de continuer.
class PhotoPickerScreen extends StatefulWidget {
  const PhotoPickerScreen({super.key});

  @override
  State<PhotoPickerScreen> createState() => _PhotoPickerScreenState();
}

class _PhotoPickerScreenState extends State<PhotoPickerScreen> {
  final ImagePicker _picker = ImagePicker();
  final List<XFile> _photos = [];

  Future<void> _pickPhotos() async {
    final picked = await _picker.pickMultiImage(limit: _maxPhotos);
    if (picked.isEmpty) return;

    final remainingSlots = _maxPhotos - _photos.length;
    final overflow = picked.length > remainingSlots;

    setState(() {
      _photos.addAll(picked.take(remainingSlots));
    });

    if (overflow) _showMaxPhotosNotice();
  }

  void _showMaxPhotosNotice() {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Maximum $_maxPhotos photos par publication.')),
    );
  }

  void _removePhoto(int index) {
    setState(() => _photos.removeAt(index));
  }

  void _continue() {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => PriceScreen(photos: List.of(_photos))),
    );
  }

  @override
  Widget build(BuildContext context) {
    final canContinue = _photos.length >= _minPhotos;

    return Scaffold(
      appBar: AppBar(title: const Text('Vos photos')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Ajoutez entre $_minPhotos et $_maxPhotos photos du produit.',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _photos.isEmpty
                  ? const _EmptyPhotosPlaceholder()
                  : GridView.builder(
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 3,
                        crossAxisSpacing: 8,
                        mainAxisSpacing: 8,
                      ),
                      itemCount: _photos.length,
                      itemBuilder: (context, index) => _PhotoTile(
                        photo: _photos[index],
                        onRemove: () => _removePhoto(index),
                      ),
                    ),
            ),
            const SizedBox(height: 16),
            OutlinedButton.icon(
              onPressed: _photos.length >= _maxPhotos ? null : _pickPhotos,
              icon: const Icon(Icons.add_photo_alternate_outlined),
              label: Text(_photos.isEmpty ? 'Choisir des photos' : 'Ajouter des photos'),
            ),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: canContinue ? _continue : null,
              child: const Text('Continuer'),
            ),
          ],
        ),
      ),
    );
  }
}

class _EmptyPhotosPlaceholder extends StatelessWidget {
  const _EmptyPhotosPlaceholder();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            Icons.image_outlined,
            size: 40,
            color: Theme.of(context).textTheme.bodyMedium?.color?.withValues(alpha: 0.5),
          ),
          const SizedBox(height: 8),
          Text(
            'Aucune photo sélectionnée',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
      ),
    );
  }
}

class _PhotoTile extends StatelessWidget {
  const _PhotoTile({required this.photo, required this.onRemove});

  final XFile photo;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Stack(
      fit: StackFit.expand,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(12),
          child: Image.file(File(photo.path), fit: BoxFit.cover),
        ),
        Positioned(
          top: 4,
          right: 4,
          child: GestureDetector(
            onTap: onRemove,
            child: const CircleAvatar(
              radius: 12,
              backgroundColor: Colors.black54,
              child: Icon(Icons.close, size: 16, color: Colors.white),
            ),
          ),
        ),
      ],
    );
  }
}
