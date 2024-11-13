import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:online_boutique/features/cart/cart_checkout_page.dart';
import 'package:online_boutique/features/cart/cart_providers.dart';
import 'package:online_boutique/features/products/products_providers.dart';

class CartPage extends HookConsumerWidget {
  const CartPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(productsProvider).asData?.value ?? [];
    final cartCount = ref.watch(cartCountProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Cart'),
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Expanded(
                child: ref.watch(cartProvider).when(
                      data: (cart) => ListView.builder(
                        itemCount: cart.items.length,
                        itemBuilder: (context, index) {
                          final item = cart.items[index];
                          final product = products
                              .firstWhere((p) => p.id == item.productId);
                          return Padding(
                            padding: const EdgeInsets.symmetric(vertical: 8.0),
                            child: ListTile(
                              leading: Image.network(
                                product.picture,
                                fit: BoxFit.cover,
                              ),
                              title: Text(product.name),
                              trailing: Text('${item.quantity}'),
                            ),
                          );
                        },
                      ),
                      error: (error, stackTrace) =>
                          Center(child: Text(error.toString())),
                      loading: () =>
                          const Center(child: CircularProgressIndicator()),
                    ),
              ),
              Column(
                children: [
                  Center(
                    child: cartCount > 0
                        ? Text('You have $cartCount items in your cart.')
                        : const Text('Your cart is empty.'),
                  ),
                  const SizedBox(height: 8),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16.0),
                    child: Row(
                      children: [
                        Expanded(
                          child: ElevatedButton(
                            onPressed: () =>
                                ref.read(cartProvider.notifier).emptyCart(),
                            style: ElevatedButton.styleFrom(
                              backgroundColor:
                                  Theme.of(context).colorScheme.error,
                              foregroundColor:
                                  Theme.of(context).colorScheme.onError,
                            ),
                            child: const Text('Empty Cart'),
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: ElevatedButton(
                            onPressed: () => Navigator.push(
                              context,
                              MaterialPageRoute(
                                  builder: (context) => CartCheckoutPage()),
                            ),
                            child: const Text('Checkout'),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
