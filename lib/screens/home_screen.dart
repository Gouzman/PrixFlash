import 'package:flutter/material.dart';

import 'photo_picker_screen.dart';

/// Écran d'accueil — point d'entrée du flux à quatre étapes du MVP :
/// photos → prix → génération → résultat. Les étapes elles-mêmes sont les
/// tickets FL-02, FL-03, FL-04, FL-05 du backlog.
class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key, required this.onToggleTheme});

  final VoidCallback onToggleTheme;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Pubprix'),
        actions: [
          IconButton(
            icon: const Icon(Icons.brightness_6_outlined),
            onPressed: onToggleTheme,
            tooltip: 'Changer de thème',
          ),
        ],
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.photo_library_outlined, size: 48),
              const SizedBox(height: 16),
              Text(
                'Sélectionnez vos photos',
                style: Theme.of(context).textTheme.headlineSmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                '2 à 6 photos du même produit, sous plusieurs angles.',
                style: Theme.of(context).textTheme.bodyMedium,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),
              ElevatedButton(
                onPressed: () {
                  Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const PhotoPickerScreen()),
                  );
                },
                child: const Text('Choisir des photos'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
