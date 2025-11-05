import 'dart:typed_data';
import 'package:image_gallery_saver/image_gallery_saver.dart';
import 'package:permission_handler/permission_handler.dart';

class SaveToGallery {
  static Future<bool> save(Uint8List imageBytes) async {
    // Vérifie et demande la permission
    var status = await Permission.storage.status;
    if (!status.isGranted) {
      status = await Permission.storage.request();
      if (!status.isGranted) {
        return false;
      }
    }

    // Enregistrement dans la galerie
    final result = await ImageGallerySaver.saveImage(
      imageBytes,
      name: "PrixFlash_${DateTime.now().millisecondsSinceEpoch}",
      quality: 100,
    );

    return result['isSuccess'] ?? false;
  }
}
