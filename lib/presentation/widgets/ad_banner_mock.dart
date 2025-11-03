import 'package:flutter/material.dart';

class AdBannerMock extends StatelessWidget {
  const AdBannerMock({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 60,
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.grey[300],
        borderRadius: BorderRadius.circular(12),
      ),
      alignment: Alignment.center,
      child: const Text(
        "Espace publicitaire (mock)",
        style: TextStyle(fontSize: 16, color: Colors.black54),
      ),
    );
  }
}
