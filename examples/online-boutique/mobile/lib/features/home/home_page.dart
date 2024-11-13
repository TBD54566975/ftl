import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:online_boutique/features/cart/cart_button.dart';
import 'package:online_boutique/features/products/product_page.dart';
import 'package:online_boutique/features/products/products_providers.dart';
import 'package:online_boutique/utils/money_utils.dart';

class HomePage extends HookConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(productsProvider);

    if (products is AsyncLoading) {
      return const Center(
        child: CircularProgressIndicator(),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Product Catalog'),
        actions: const [CartButton()],
      ),
      body: ref.watch(productsProvider).when(
            data: (products) => GridView.builder(
              padding: EdgeInsets.only(
                left: 8,
                right: 8,
                top: 8,
                bottom: MediaQuery.of(context).padding.bottom + 8,
              ),
              itemCount: products.length,
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                childAspectRatio: 2 / 3,
                crossAxisSpacing: 8,
                mainAxisSpacing: 8,
              ),
              itemBuilder: (ctx, i) {
                final product = products[i];
                return GestureDetector(
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => ProductPage(product: product),
                      ),
                    );
                  },
                  child: Card(
                    child: Column(
                      children: [
                        Expanded(
                          child: Image.network(
                            product.picture,
                            fit: BoxFit.cover,
                          ),
                        ),
                        const SizedBox(height: 10),
                        Text(
                          product.name,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 5),
                        Text(
                          '\$${fromMoney(product.priceUsd)}',
                          style: const TextStyle(
                            fontSize: 14,
                            color: Colors.grey,
                          ),
                        ),
                        const SizedBox(height: 10),
                      ],
                    ),
                  ),
                );
              },
            ),
            error: (error, stack) => Padding(
              padding: const EdgeInsets.all(16),
              child: Center(
                  child: Text(
                'Error: $error',
                textAlign: TextAlign.center,
                style: Theme.of(context)
                    .textTheme
                    .bodyLarge
                    ?.copyWith(color: Theme.of(context).colorScheme.error),
              )),
            ),
            loading: () => const Center(
              child: CircularProgressIndicator(),
            ),
          ),
    );
  }
}
