import 'dart:convert';
import 'package:http/http.dart' as http;

isOkResponse(int statusCode) {
  return statusCode >= 200 && statusCode < 300;
}

class FTLHttpClient {
  final String baseUrl;
  final http.Client httpClient;

  FTLHttpClient._({required this.baseUrl, required this.httpClient});

  static FTLHttpClient? _instance;

  static void initialize(
      {required String baseUrl, required http.Client httpClient}) {
    _instance = FTLHttpClient._(baseUrl: baseUrl, httpClient: httpClient);
  }

  static FTLHttpClient get instance {
    assert(_instance != null, 'FTLHttpClient must be initialized first');
    return _instance!;
  }

  Future<http.Response> get(
    String path, {
    String? requestJson,
    Map<String, String>? headers,
  }) {
    Uri uri;
    if (requestJson == null || requestJson.isEmpty) {
      uri = Uri.http("localhost:8891", path);
    } else {
      uri = Uri.http("localhost:8891", path, {'@json': requestJson});
    }
    return httpClient.get(uri, headers: headers);
  }

  Future<http.Response> post(
    String path, {
    Map<String, dynamic>? request,
    Map<String, String>? headers,
  }) {
    return httpClient.post(
      Uri.http(baseUrl, path),
      body: json.encode(request),
      headers: headers,
    );
  }
}
