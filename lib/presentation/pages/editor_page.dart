import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../viewmodels/price_tag_viewmodel.dart';
import '../widgets/price_tag_widget.dart';
import '../widgets/design_selector.dart';
import '../widgets/font_selector.dart';
import '../widgets/logo_uploader.dart';

class EditorPage extends ConsumerStatefulWidget {
  const EditorPage({super.key});

  @override
  ConsumerState<EditorPage> createState() => _EditorPageState();
}

class _EditorPageState extends ConsumerState<EditorPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final tag = ref.watch(priceTagProvider);
    final viewModel = ref.read(priceTagProvider.notifier);

    return Scaffold(
      appBar: AppBar(
        title: const Text("Créer une étiquette"),
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          labelColor: Theme.of(context).colorScheme.primary,
          unselectedLabelColor: Colors.grey,
          indicatorColor: Theme.of(context).colorScheme.primary,
          tabs: const [
            Tab(text: "Texte"),
            Tab(text: "Couleurs"),
            Tab(text: "Design"),
            Tab(text: "Logo & Police"),
          ],
        ),
      ),
      body: Column(
        children: [
          // ✅ Étiquette toujours visible en haut
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Center(child: PriceTagWidget(tag: tag)),
          ),

          // ✅ Onglets de configuration
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                // 📝 Onglet 1 : Texte & Prix
                _buildTextSection(viewModel),
                // 🎨 Onglet 2 : Couleurs
                _buildColorSection(viewModel, tag),
                // 🧩 Onglet 3 : Modèles
                const DesignSelector(),
                // 🖼️ Onglet 4 : Logo & Police
                _buildFontLogoSection(),
              ],
            ),
          ),
        ],
      ),

      // ✅ Bouton “Voir l’aperçu” flottant (fixe en bas)
      floatingActionButtonLocation: FloatingActionButtonLocation.centerFloat,
      floatingActionButton: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16.0),
        child: ElevatedButton.icon(
          icon: const Icon(Icons.preview),
          label: const Text("Voir l’aperçu"),
          onPressed: () => context.push('/preview'),
          style: ElevatedButton.styleFrom(
            minimumSize: const Size(double.infinity, 55),
            backgroundColor: Theme.of(context).colorScheme.primary,
            foregroundColor: Colors.white,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
            ),
          ),
        ),
      ),
    );
  }

  // SECTION 1 — Texte & Prix
  Widget _buildTextSection(viewModel) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: ListView(
        children: [
          TextField(
            decoration: const InputDecoration(
              labelText: "Nom du produit",
              border: OutlineInputBorder(),
            ),
            onChanged: viewModel.updateProductName,
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: TextField(
                  decoration: const InputDecoration(
                    labelText: "Ancien prix",
                    border: OutlineInputBorder(),
                  ),
                  keyboardType: TextInputType.number,
                  onChanged: viewModel.updateOldPrice,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: TextField(
                  decoration: const InputDecoration(
                    labelText: "Nouveau prix",
                    border: OutlineInputBorder(),
                  ),
                  keyboardType: TextInputType.number,
                  onChanged: viewModel.updateNewPrice,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // SECTION 2 — Couleurs
  Widget _buildColorSection(viewModel, tag) {
    final bgColors = ['#000000', '#ff9800', '#ffffff', '#4caf50', '#2196f3'];
    final textColors = ['#ffffff', '#000000', '#f44336', '#ffeb3b', '#00bcd4'];

    return Padding(
      padding: const EdgeInsets.all(16),
      child: ListView(
        children: [
          const Text("🎨 Couleur de fond"),
          const SizedBox(height: 10),
          Wrap(
            spacing: 10,
            children: bgColors.map((hex) {
              final color = Color(int.parse(hex.replaceAll('#', '0xff')));
              return GestureDetector(
                onTap: () => viewModel.updateBgColor(hex),
                child: CircleAvatar(
                  backgroundColor: color,
                  radius: 18,
                  child: tag.backgroundColor == hex
                      ? const Icon(Icons.check, color: Colors.white)
                      : null,
                ),
              );
            }).toList(),
          ),
          const SizedBox(height: 24),
          const Text("🖋️ Couleur du texte"),
          const SizedBox(height: 10),
          Wrap(
            spacing: 10,
            children: textColors.map((hex) {
              final color = Color(int.parse(hex.replaceAll('#', '0xff')));
              return GestureDetector(
                onTap: () => viewModel.updateTextColor(hex),
                child: CircleAvatar(
                  backgroundColor: color,
                  radius: 18,
                  child: tag.textColor == hex
                      ? const Icon(Icons.check, color: Colors.black)
                      : null,
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  // SECTION 4 — Logo & Police
  Widget _buildFontLogoSection() {
    return const Padding(
      padding: EdgeInsets.all(16),
      child: SingleChildScrollView(
        child: Column(
          children: [
            FontSelector(),
            LogoUploader(),
          ],
        ),
      ),
    );
  }
}
