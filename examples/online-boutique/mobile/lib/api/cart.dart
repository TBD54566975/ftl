// ignore_for_file: unused_import
library cart;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
import 'builtin.dart' as builtin;


class AddItemRequest{
  String userId;
  Item item;

  AddItemRequest({  required this.userId,  required this.item,  });

  Map<String, dynamic> toJson() {
    return {
      'userId': ((dynamic v) => v)(userId),
      'item': ((dynamic v) => v.toJson())(item),
    };
  }

  factory AddItemRequest.fromJson(Map<String, dynamic> map) {
    return AddItemRequest(
      userId: ((dynamic v) => v)(map['userId']), item: ((dynamic v) => Item.fromJson(v))(map['item']), 
    );
  }
}

class AddItemResponse{

  AddItemResponse();

  Map<String, dynamic> toJson() {
    return {
    };
  }

  factory AddItemResponse.fromJson(Map<String, dynamic> map) {
    return AddItemResponse(
      
    );
  }
}

class Cart{
  String userId;
  List<Item> items;

  Cart({  required this.userId,  required this.items,  });

  Map<String, dynamic> toJson() {
    return {
      'userId': ((dynamic v) => v)(userId),
      'items': ((dynamic v) => v.map((v) => v.toJson()).cast<Item>().toList())(items),
    };
  }

  factory Cart.fromJson(Map<String, dynamic> map) {
    return Cart(
      userId: ((dynamic v) => v)(map['userId']), items: ((dynamic v) => v.map((v) => Item.fromJson(v)).cast<Item>().toList())(map['items']), 
    );
  }
}

class EmptyCartRequest{
  String userId;

  EmptyCartRequest({  required this.userId,  });

  Map<String, dynamic> toJson() {
    return {
      'userId': ((dynamic v) => v)(userId),
    };
  }

  factory EmptyCartRequest.fromJson(Map<String, dynamic> map) {
    return EmptyCartRequest(
      userId: ((dynamic v) => v)(map['userId']), 
    );
  }
}

class EmptyCartResponse{

  EmptyCartResponse();

  Map<String, dynamic> toJson() {
    return {
    };
  }

  factory EmptyCartResponse.fromJson(Map<String, dynamic> map) {
    return EmptyCartResponse(
      
    );
  }
}

class GetCartRequest{
  String userId;

  GetCartRequest({  required this.userId,  });

  Map<String, dynamic> toJson() {
    return {
      'userId': ((dynamic v) => v)(userId),
    };
  }

  factory GetCartRequest.fromJson(Map<String, dynamic> map) {
    return GetCartRequest(
      userId: ((dynamic v) => v)(map['userId']), 
    );
  }
}

class Item{
  String productId;
  int quantity;

  Item({  required this.productId,  required this.quantity,  });

  Map<String, dynamic> toJson() {
    return {
      'productId': ((dynamic v) => v)(productId),
      'quantity': ((dynamic v) => v)(quantity),
    };
  }

  factory Item.fromJson(Map<String, dynamic> map) {
    return Item(
      productId: ((dynamic v) => v)(map['productId']), quantity: ((dynamic v) => v)(map['quantity']), 
    );
  }
}


class CartClient {
  final FTLHttpClient ftlClient;

  CartClient({required this.ftlClient});


  Future<AddItemResponse> addItem(
    AddItemRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.post('/cart/add', request: request.toJson());
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return AddItemResponse.fromJson(body);
    } else {
      throw Exception('Failed to get addItem response');
    }
  }

  Future<EmptyCartResponse> emptyCart(
    EmptyCartRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.post('/cart/empty', request: request.toJson());
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return EmptyCartResponse.fromJson(body);
    } else {
      throw Exception('Failed to get emptyCart response');
    }
  }

  Future<Cart> getCart(
    GetCartRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.get(
      '/cart', 
      requestJson: json.encode(request.toJson()),
      headers: headers,
    );
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return Cart.fromJson(body);
    } else {
      throw Exception('Failed to get getCart response');
    }
  }
}
