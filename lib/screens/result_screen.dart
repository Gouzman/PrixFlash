import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:gal/gal.dart';
import 'package:http/http.dart' as http;
import 'package:share_plus/share_plus.dart';

import '../models/generation_result.dart';

/// Écran de résultat — FL-05.
///
/// Affiche le visuel généré et propose de le télécharger dans la galerie
/// ou de le partager vers WhatsApp / Facebook / Instagram.
class ResultScreen extends StatefulWidget {
  const ResultScreen({super.key, required this.result});

  final GenerationResult result;

  @override
  State<ResultScreen> createState() => _ResultScreenState();
}

class _ResultScreenState extends State<ResultScreen> {
  Uint8List? _imageBytes;
  bool _isBusy = false;

  Future<Uint8List> _loadImageBytes() async {
    if (_imageBytes != null) return _imageBytes!;
    final response = await http.get(Uri.parse(widget.result.imageUrl!));
    if (response.statusCode != 200) {
      throw Exception('Impossible de récupérer le visuel généré.');
    }
    _imageBytes = response.bodyBytes;
    return _imageBytes!;
  }

  Future<void> _download() async {
    setState(() => _isBusy = true);
    try {
      final bytes = await _loadImageBytes();
      await Gal.putImageBytes(bytes, name: 'pubprix_${DateTime.now().millisecondsSinceEpoch}');
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Publication enregistrée dans la galerie.')),
      );
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Le téléchargement a échoué. Réessayez.')),
      );
    } finally {
      if (mounted) setState(() => _isBusy = false);
    }
  }

  // share_plus ouvre la feuille de partage native : l'utilisateur y choisit
  // lui-même WhatsApp, Facebook ou Instagram. Il n'existe pas d'API publique
  // pour cibler directement une app tierce sur tous les OS.
  Future<void> _share(String platformLabel) async {
    setState(() => _isBusy = true);
    try {
      final bytes = await _loadImageBytes();
      await Share.shareXFiles(
        [XFile.fromData(bytes, mimeType: 'image/png', name: 'pubprix.png')],
        text: 'Ma publication Pubprix, prête pour $platformLabel.',
      );
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Le partage a échoué. Réessayez.')),
      );
    } finally {
      if (mounted) setState(() => _isBusy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final imageUrl = widget.result.imageUrl;

    return Scaffold(
      appBar: AppBar(title: const Text('Votre publication')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Expanded(
              child: imageUrl == null
                  ? const Center(child: Text('Aucun visuel disponible.'))
                  : ClipRRect(
                      borderRadius: BorderRadius.circular(16),
                      child: Image.network(
                        imageUrl,
                        fit: BoxFit.contain,
                        loadingBuilder: (context, child, progress) {
                          if (progress == null) return child;
                          return const Center(child: CircularProgressIndicator());
                        },
                        errorBuilder: (context, error, stackTrace) =>
                            const Center(child: Text('Impossible d\'afficher le visuel.')),
                      ),
                    ),
            ),
            const SizedBox(height: 16),
            OutlinedButton.icon(
              onPressed: imageUrl == null || _isBusy ? null : _download,
              icon: const Icon(Icons.download_outlined),
              label: const Text('Télécharger'),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              alignment: WrapAlignment.center,
              children: [
                _ShareButton(
                  label: 'WhatsApp',
                  icon: Icons.chat_outlined,
                  onPressed: imageUrl == null || _isBusy ? null : () => _share('WhatsApp'),
                ),
                _ShareButton(
                  label: 'Facebook',
                  icon: Icons.facebook_outlined,
                  onPressed: imageUrl == null || _isBusy ? null : () => _share('Facebook'),
                ),
                _ShareButton(
                  label: 'Instagram',
                  icon: Icons.camera_alt_outlined,
                  onPressed: imageUrl == null || _isBusy ? null : () => _share('Instagram'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ShareButton extends StatelessWidget {
  const _ShareButton({required this.label, required this.icon, required this.onPressed});

  final String label;
  final IconData icon;
  final VoidCallback? onPressed;

  @override
  Widget build(BuildContext context) {
    return ElevatedButton.icon(
      onPressed: onPressed,
      icon: Icon(icon, size: 18),
      label: Text(label),
    );
  }
}
