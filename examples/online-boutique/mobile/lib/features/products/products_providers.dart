import 'dart:async';

import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:online_boutique/api/productcatalog.dart';
import 'package:online_boutique/utils/api_providers.dart';

final productsProvider = AsyncNotifierProvider<ProductsNotifier, List<Product>>(
    () => ProductsNotifier());

class ProductsNotifier extends AsyncNotifier<List<Product>> {
  Future<List<Product>> refreshProducts() async {
    final products =
        await ref.read(productsCatalogProvider).list(ListRequest());

    state = AsyncData(products.products);
    return products.products;
  }

  @override
  FutureOr<List<Product>> build() {
    return refreshProducts();
  }
}
