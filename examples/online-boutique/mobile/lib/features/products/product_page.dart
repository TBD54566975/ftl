import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:online_boutique/api/productcatalog.dart';
import 'package:online_boutique/features/cart/cart_button.dart';
import 'package:online_boutique/features/cart/cart_providers.dart';

class ProductPage extends HookConsumerWidget {
  final Product product;
  const ProductPage({required this.product, super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isAdding = useState(false);
    final quantity = useState(1);

    return Scaffold(
      appBar: AppBar(
        title: Text(product.name),
        actions: const [CartButton()],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Expanded(
                child: SingleChildScrollView(
                  child: Column(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(10),
                        child: Image.network(
                          product.picture,
                          fit: BoxFit.cover,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Column(
                        children: [
                          Text(
                            'SKU: ${product.id}',
                            style: Theme.of(context).textTheme.labelMedium,
                            textAlign: TextAlign.center,
                          ),
                          const SizedBox(height: 16),
                          Text(
                            product.description,
                            style: Theme.of(context).textTheme.titleLarge,
                            textAlign: TextAlign.center,
                          ),
                          const SizedBox(height: 8),
                          Wrap(
                            spacing: 8.0,
                            children: product.categories.map((category) {
                              return Chip(label: Text(category));
                            }).toList(),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('Quantity:'),
                  Row(
                    children: [
                      IconButton(
                        onPressed: () {
                          if (quantity.value > 1) {
                            quantity.value--;
                          }
                        },
                        icon: const Icon(Icons.remove),
                      ),
                      Text(quantity.value.toString()),
                      IconButton(
                        onPressed: () {
                          quantity.value++;
                        },
                        icon: const Icon(Icons.add),
                      ),
                    ],
                  ),
                ],
              ),
              ElevatedButton(
                onPressed: () {
                  isAdding.value = true;
                  ref
                      .read(cartProvider.notifier)
                      .addItem(
                        productId: product.id,
                        quantity: quantity.value,
                      )
                      .then((value) {
                    isAdding.value = false;

                    const snackBar = SnackBar(content: Text('Added to cart!'));
                    ScaffoldMessenger.of(context).showSnackBar(snackBar);
                  });
                },
                child: const Text('Add to Cart'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
