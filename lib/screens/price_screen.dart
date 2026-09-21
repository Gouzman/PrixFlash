import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../services/pubprix_api.dart';
import 'result_screen.dart';

const String _currency = 'FCFA';

/// Écran de saisie prix/nom — FL-03.
///
/// Déclenche aussi l'appel API de génération (FL-04) : création du
/// brouillon, upload des photos, mise à jour prix/nom, puis génération du
/// visuel.
class PriceScreen extends StatefulWidget {
  const PriceScreen({super.key, required this.photos});

  final List<XFile> photos;

  @override
  State<PriceScreen> createState() => _PriceScreenState();
}

class _PriceScreenState extends State<PriceScreen> {
  final _formKey = GlobalKey<FormState>();
  final _priceController = TextEditingController();
  final _nameController = TextEditingController();
  final _api = PubprixApiService();

  bool _isSubmitting = false;

  @override
  void dispose() {
    _priceController.dispose();
    _nameController.dispose();
    _api.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;

    setState(() => _isSubmitting = true);

    final price = double.parse(_priceController.text.replaceAll(',', '.'));
    final name = _nameController.text.trim();

    try {
      final draft = await _api.createDraft();
      await _api.uploadPhotos(draft.id, photos: widget.photos);
      await _api.updateDraft(draft.id, price: price, name: name.isEmpty ? null : name);
      final result = await _api.generate(draft.id);

      if (!mounted) return;
      Navigator.of(context).push(
        MaterialPageRoute(builder: (_) => ResultScreen(result: result)),
      );
    } on PubprixApiException catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(error.message)));
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Une erreur est survenue. Réessayez.')),
      );
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Prix & nom')),
      body: AbsorbPointer(
        absorbing: _isSubmitting,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextFormField(
                  controller: _priceController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: const InputDecoration(
                    labelText: 'Prix',
                    suffixText: _currency,
                    border: OutlineInputBorder(),
                  ),
                  validator: (value) {
                    final trimmed = value?.trim() ?? '';
                    if (trimmed.isEmpty) return 'Le prix est obligatoire.';
                    final parsed = double.tryParse(trimmed.replaceAll(',', '.'));
                    if (parsed == null || parsed <= 0) return 'Entrez un prix valide.';
                    return null;
                  },
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _nameController,
                  decoration: const InputDecoration(
                    labelText: 'Nom du produit (optionnel)',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 24),
                ElevatedButton(
                  onPressed: _isSubmitting ? null : _submit,
                  child: _isSubmitting
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                        )
                      : const Text('Générer ma publication'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
