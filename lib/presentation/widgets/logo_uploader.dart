import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../viewmodels/price_tag_viewmodel.dart';

class LogoUploader extends ConsumerWidget {
  const LogoUploader({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tag = ref.watch(priceTagProvider);
    final viewModel = ref.read(priceTagProvider.notifier);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 24),
        const Text("🖼️ Logo du commerçant"),
        const SizedBox(height: 10),
        Row(
          children: [
            if (tag.logoPath != null)
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: Image.file(
                  File(tag.logoPath!),
                  width: 60,
                  height: 60,
                  fit: BoxFit.cover,
                ),
              ),
            const SizedBox(width: 10),
            ElevatedButton.icon(
              icon: const Icon(Icons.upload),
              label: const Text("Uploader un logo"),
              onPressed: () async {
                final picker = ImagePicker();
                final file = await picker.pickImage(source: ImageSource.gallery);
                if (file != null) viewModel.updateLogoPath(file.path);
              },
            ),
          ],
        ),
      ],
    );
  }
}
