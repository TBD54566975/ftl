import 'dart:async';

import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:online_boutique/api/cart.dart';
import 'package:online_boutique/api/ftl_client.dart';

final cartProvider =
    AsyncNotifierProvider<CartNotifier, Cart>(() => CartNotifier());

final cartCountProvider = Provider<int>((ref) =>
    ref
        .watch(cartProvider)
        .asData
        ?.value
        .items
        .fold(0, (sum, item) => (sum ?? 0) + item.quantity) ??
    0);

class CartNotifier extends AsyncNotifier<Cart> {
  @override
  FutureOr<Cart> build() async {
    return getCart();
  }

  Future<void> addItem({
    required String productId,
    required int quantity,
  }) async {
    CartClient(ftlClient: FTLHttpClient.instance)
        .addItem(AddItemRequest(
          userId: 'a',
          item: Item(productId: productId, quantity: quantity),
        ))
        .then((value) => getCart());
  }

  Future<Cart> getCart() async {
    final cart = await CartClient(ftlClient: FTLHttpClient.instance)
        .getCart(GetCartRequest(userId: 'a'));
    state = AsyncData(cart);
    return cart;
  }

  Future<void> emptyCart() async {
    CartClient(ftlClient: FTLHttpClient.instance)
        .emptyCart(EmptyCartRequest(userId: 'a'))
        .then((value) => getCart());
  }
}
