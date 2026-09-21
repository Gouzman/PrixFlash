import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart';

import '../config/app_config.dart';
import '../models/draft.dart';
import '../models/generation_result.dart';

/// Client HTTP pour l'API Pubprix — FL-04.
///
/// Enchaîne les quatre appels du flux de génération : création du
/// brouillon, upload des photos, mise à jour prix/nom, puis génération du
/// visuel.
class PubprixApiService {
  PubprixApiService({http.Client? client, String? baseUrl})
    : _client = client ?? http.Client(),
      _baseUrl = baseUrl ?? AppConfig.apiBaseUrl;

  final http.Client _client;
  final String _baseUrl;

  /// POST /drafts — API-02.
  Future<Draft> createDraft() async {
    final response = await _client.post(
      Uri.parse('$_baseUrl/drafts'),
      headers: {'Content-Type': 'application/json'},
    );
    _throwIfNotOk(response, 'la création du brouillon');
    return Draft.fromJson(jsonDecode(response.body) as Map<String, dynamic>);
  }

  /// POST /drafts/{id}/photos — API-03.
  ///
  /// Upload multipart des photos sélectionnées ; l'API les stocke sur R2 et
  /// renvoie leurs URLs.
  Future<List<String>> uploadPhotos(String draftId, {required List<XFile> photos}) async {
    final request = http.MultipartRequest(
      'POST',
      Uri.parse('$_baseUrl/drafts/$draftId/photos'),
    );
    for (final photo in photos) {
      request.files.add(await http.MultipartFile.fromPath('photos', photo.path));
    }

    final streamedResponse = await _client.send(request);
    final response = await http.Response.fromStream(streamedResponse);
    _throwIfNotOk(response, "l'upload des photos");

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return (body['photos'] as List<dynamic>).cast<String>();
  }

  /// PATCH /drafts/{id} — API-04.
  Future<Draft> updateDraft(String draftId, {required double price, String? name}) async {
    final response = await _client.patch(
      Uri.parse('$_baseUrl/drafts/$draftId'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'price': price,
        if (name != null && name.isNotEmpty) 'name': name,
      }),
    );
    _throwIfNotOk(response, 'la mise à jour du brouillon');
    return Draft.fromJson(jsonDecode(response.body) as Map<String, dynamic>);
  }

  /// POST /drafts/{id}/generate — API-05.
  ///
  /// Aucune photo dans le corps de la requête : l'API les a déjà (upload
  /// séparé via [uploadPhotos]).
  Future<GenerationResult> generate(String draftId) async {
    final response = await _client.post(
      Uri.parse('$_baseUrl/drafts/$draftId/generate'),
      headers: {'Content-Type': 'application/json'},
    );
    _throwIfNotOk(response, 'la génération de la publication');
    return GenerationResult.fromJson(jsonDecode(response.body) as Map<String, dynamic>);
  }

  void _throwIfNotOk(http.Response response, String action) {
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw PubprixApiException(
        'Échec pendant $action (${response.statusCode}).',
      );
    }
  }

  void dispose() => _client.close();
}

class PubprixApiException implements Exception {
  PubprixApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
