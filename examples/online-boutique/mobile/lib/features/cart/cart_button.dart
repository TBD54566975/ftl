import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:online_boutique/features/cart/cart_page.dart';
import 'package:online_boutique/features/cart/cart_providers.dart';

class CartButton extends HookConsumerWidget {
  const CartButton({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cartCount = ref.watch(cartCountProvider);

    return IconButton(
      icon: cartCount > 0
          ? Badge(
              label: Text('$cartCount'),
              child: const Icon(Icons.shopping_cart),
            )
          : const Icon(Icons.shopping_cart),
      onPressed: () {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (context) => const CartPage(),
          ),
        );
      },
    );
  }
}
