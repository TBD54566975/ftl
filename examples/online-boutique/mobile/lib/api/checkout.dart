// ignore_for_file: unused_import
library checkout;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
import 'builtin.dart' as builtin;
import 'cart.dart' as cart;
import 'currency.dart' as currency;
import 'payment.dart' as payment;
import 'productcatalog.dart' as productcatalog;
import 'shipping.dart' as shipping;


class ErrorResponse{
  String message;

  ErrorResponse({  required this.message,  });

  Map<String, dynamic> toJson() {
    return {
      'message': ((dynamic v) => v)(message),
    };
  }

  factory ErrorResponse.fromJson(Map<String, dynamic> map) {
    return ErrorResponse(
      message: ((dynamic v) => v)(map['message']), 
    );
  }
}

class Order{
  String id;
  String shippingTrackingId;
  currency.Money shippingCost;
  shipping.Address shippingAddress;
  List<OrderItem> items;

  Order({  required this.id,  required this.shippingTrackingId,  required this.shippingCost,  required this.shippingAddress,  required this.items,  });

  Map<String, dynamic> toJson() {
    return {
      'id': ((dynamic v) => v)(id),
      'shippingTrackingId': ((dynamic v) => v)(shippingTrackingId),
      'shippingCost': ((dynamic v) => v.toJson())(shippingCost),
      'shippingAddress': ((dynamic v) => v.toJson())(shippingAddress),
      'items': ((dynamic v) => v.map((v) => v.toJson()).cast<OrderItem>().toList())(items),
    };
  }

  factory Order.fromJson(Map<String, dynamic> map) {
    return Order(
      id: ((dynamic v) => v)(map['id']), shippingTrackingId: ((dynamic v) => v)(map['shippingTrackingId']), shippingCost: ((dynamic v) => currency.Money.fromJson(v))(map['shippingCost']), shippingAddress: ((dynamic v) => shipping.Address.fromJson(v))(map['shippingAddress']), items: ((dynamic v) => v.map((v) => OrderItem.fromJson(v)).cast<OrderItem>().toList())(map['items']), 
    );
  }
}

class OrderItem{
  cart.Item item;
  currency.Money cost;

  OrderItem({  required this.item,  required this.cost,  });

  Map<String, dynamic> toJson() {
    return {
      'item': ((dynamic v) => v.toJson())(item),
      'cost': ((dynamic v) => v.toJson())(cost),
    };
  }

  factory OrderItem.fromJson(Map<String, dynamic> map) {
    return OrderItem(
      item: ((dynamic v) => cart.Item.fromJson(v))(map['item']), cost: ((dynamic v) => currency.Money.fromJson(v))(map['cost']), 
    );
  }
}

class PlaceOrderRequest{
  String userId;
  String userCurrency;
  shipping.Address address;
  String email;
  payment.CreditCardInfo creditCard;

  PlaceOrderRequest({  required this.userId,  required this.userCurrency,  required this.address,  required this.email,  required this.creditCard,  });

  Map<String, dynamic> toJson() {
    return {
      'userId': ((dynamic v) => v)(userId),
      'userCurrency': ((dynamic v) => v)(userCurrency),
      'address': ((dynamic v) => v.toJson())(address),
      'email': ((dynamic v) => v)(email),
      'creditCard': ((dynamic v) => v.toJson())(creditCard),
    };
  }

  factory PlaceOrderRequest.fromJson(Map<String, dynamic> map) {
    return PlaceOrderRequest(
      userId: ((dynamic v) => v)(map['userId']), userCurrency: ((dynamic v) => v)(map['userCurrency']), address: ((dynamic v) => shipping.Address.fromJson(v))(map['address']), email: ((dynamic v) => v)(map['email']), creditCard: ((dynamic v) => payment.CreditCardInfo.fromJson(v))(map['creditCard']), 
    );
  }
}


class CheckoutClient {
  final FTLHttpClient ftlClient;

  CheckoutClient({required this.ftlClient});


  Future<Order> placeOrder(
    PlaceOrderRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.post('/checkout/userId', request: request.toJson());
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return Order.fromJson(body);
    } else {
      throw Exception('Failed to get placeOrder response');
    }
  }

}
