import 'dart:io';

import 'package:http/http.dart' as http;
import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:online_boutique/api/ftl_client.dart';
import 'package:online_boutique/api/productcatalog.dart';

final ftlClientProvider = Provider<FTLHttpClient>((ref) {
  final host = Platform.isAndroid ? '10.0.2.2' : 'localhost:8891';

  FTLHttpClient.initialize(
    baseUrl: host,
    httpClient: http.Client(),
  );

  return FTLHttpClient.instance;
});

final productsCatalogProvider = Provider<ProductcatalogClient>((ref) {
  return ProductcatalogClient(ftlClient: ref.watch(ftlClientProvider));
});
