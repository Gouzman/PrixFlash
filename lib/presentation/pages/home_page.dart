import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../widgets/ad_banner_mock.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('PrixFlash 💥')),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            const AdBannerMock(),
            const SizedBox(height: 20),
            Expanded(
              child: GridView.count(
                crossAxisCount: 2,
                childAspectRatio: 0.9,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                children: [
                  _buildCard(
                    context,
                    'Créer une étiquette',
                    Icons.add,
                    '/editor',
                  ),
                  _buildCard(context, 'Historique', Icons.history, '/history'),
                  _buildCard(context, 'Plan Pro', Icons.star, '/plan-pro'),
                  _buildCard(context, 'À propos', Icons.info, '/about'),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCard(
    BuildContext context,
    String title,
    IconData icon,
    String route,
  ) {
    return InkWell(
      borderRadius: BorderRadius.circular(16),
      onTap: () => context.push(route),
      child: Card(
        elevation: 3,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                icon,
                size: 40,
                color: Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(height: 10),
              Text(title, style: const TextStyle(fontWeight: FontWeight.w600)),
            ],
          ),
        ),
      ),
    );
  }
}
