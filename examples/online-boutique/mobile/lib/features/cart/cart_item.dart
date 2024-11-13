class CartItem {
  String productId;
  int quantity;

  CartItem({required this.productId, required this.quantity});

  factory CartItem.fromJson(Map<String, dynamic> json) {
    return CartItem(
      productId: json['ProductID'],
      quantity: json['Quantity'],
    );
  }
}
