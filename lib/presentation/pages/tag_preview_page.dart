import 'package:flutter/material.dart';
import '../../domain/models/price_tag_model.dart';
import '../widgets/price_tag_widget.dart';

class TagPreviewPage extends StatelessWidget {
  final PriceTag tag;
  
  const TagPreviewPage({super.key, required this.tag});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[100],
      appBar: AppBar(
        title: Text('Aperçu - ${tag.productName}'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        actions: [
          IconButton(
            onPressed: () => _shareTag(context),
            icon: const Icon(Icons.share),
            tooltip: 'Partager',
          ),
          IconButton(
            onPressed: () => _downloadTag(context),
            icon: const Icon(Icons.download),
            tooltip: 'Télécharger',
          ),
        ],
      ),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Aperçu principal
              Container(
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: const [
                    BoxShadow(
                      color: Colors.black12,
                      blurRadius: 8,
                      offset: Offset(0, 4),
                    ),
                  ],
                ),
                child: PriceTagWidget(tag: tag),
              ),
              
              const SizedBox(height: 32),
              
              // Informations détaillées
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Détails de l\'étiquette',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 12),
                      _buildDetailRow('Produit', tag.productName),
                      _buildDetailRow('Ancien prix', '${tag.oldPrice.toStringAsFixed(2)} ${tag.currency}'),
                      _buildDetailRow('Nouveau prix', '${tag.newPrice.toStringAsFixed(2)} ${tag.currency}'),
                      _buildDetailRow('Réduction', '${((tag.oldPrice - tag.newPrice) / tag.oldPrice * 100).toStringAsFixed(1)}%'),
                      _buildDetailRow('Design', _getDesignName(tag.designType)),
                      _buildDetailRow('Police', tag.fontFamily),
                      _buildColorRow('Couleur fond', tag.backgroundColor),
                      _buildColorRow('Couleur texte', tag.textColor),
                      if (tag.logoPath != null)
                        _buildDetailRow('Logo', tag.logoPath!.split('/').last),
                    ],
                  ),
                ),
              ),
              
              const SizedBox(height: 24),
              
              // Boutons d'action
              Row(
                children: [
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed: () => Navigator.pop(context),
                      icon: const Icon(Icons.edit),
                      label: const Text('Modifier'),
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
                      onPressed: () => _generateNewVariation(context),
                      icon: const Icon(Icons.auto_awesome),
                      label: const Text('Varier'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Theme.of(context).colorScheme.secondary,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(color: Colors.grey),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildColorRow(String label, String colorHex) {
    final color = Color(int.parse(colorHex.replaceAll('#', '0xff')));
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
          Container(
            width: 20,
            height: 20,
            decoration: BoxDecoration(
              color: color,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: Colors.grey[300]!),
            ),
          ),
          const SizedBox(width: 8),
          Text(
            colorHex.toUpperCase(),
            style: const TextStyle(
              color: Colors.grey,
              fontFamily: 'monospace',
            ),
          ),
        ],
      ),
    );
  }

  String _getDesignName(String designType) {
    switch (designType) {
      case 'square':
        return 'Carré';
      case 'hanging':
        return 'Suspendu';
      case 'ribbon':
        return 'Ruban';
      default:
        return 'Inconnu';
    }
  }

  void _shareTag(BuildContext context) {
    // TODO: Implémenter le partage
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Fonctionnalité de partage à implémenter'),
        backgroundColor: Colors.orange,
      ),
    );
  }

  void _downloadTag(BuildContext context) {
    // TODO: Implémenter le téléchargement
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Téléchargement en cours...'),
        backgroundColor: Colors.green,
      ),
    );
  }

  void _generateNewVariation(BuildContext context) {
    // TODO: Implémenter les variations automatiques
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Génération de variations à implémenter'),
        backgroundColor: Colors.purple,
      ),
    );
  }
}