// ignore_for_file: unused_import
library shipping;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
import 'builtin.dart' as builtin;
import 'cart.dart' as cart;
import 'currency.dart' as currency;


class Address{
  String streetAddress;
  String city;
  String state;
  String country;
  int zipCode;

  Address({  required this.streetAddress,  required this.city,  required this.state,  required this.country,  required this.zipCode,  });

  Map<String, dynamic> toJson() {
    return {
      'streetAddress': ((dynamic v) => v)(streetAddress),
      'city': ((dynamic v) => v)(city),
      'state': ((dynamic v) => v)(state),
      'country': ((dynamic v) => v)(country),
      'zipCode': ((dynamic v) => v)(zipCode),
    };
  }

  factory Address.fromJson(Map<String, dynamic> map) {
    return Address(
      streetAddress: ((dynamic v) => v)(map['streetAddress']), city: ((dynamic v) => v)(map['city']), state: ((dynamic v) => v)(map['state']), country: ((dynamic v) => v)(map['country']), zipCode: ((dynamic v) => v)(map['zipCode']), 
    );
  }
}

class ShipOrderResponse{
  String id;

  ShipOrderResponse({  required this.id,  });

  Map<String, dynamic> toJson() {
    return {
      'id': ((dynamic v) => v)(id),
    };
  }

  factory ShipOrderResponse.fromJson(Map<String, dynamic> map) {
    return ShipOrderResponse(
      id: ((dynamic v) => v)(map['id']), 
    );
  }
}

class ShippingRequest{
  Address address;
  List<cart.Item> items;

  ShippingRequest({  required this.address,  required this.items,  });

  Map<String, dynamic> toJson() {
    return {
      'address': ((dynamic v) => v.toJson())(address),
      'items': ((dynamic v) => v.map((v) => v.toJson()).cast<cart.Item>().toList())(items),
    };
  }

  factory ShippingRequest.fromJson(Map<String, dynamic> map) {
    return ShippingRequest(
      address: ((dynamic v) => Address.fromJson(v))(map['address']), items: ((dynamic v) => v.map((v) => cart.Item.fromJson(v)).cast<cart.Item>().toList())(map['items']), 
    );
  }
}


class ShippingClient {
  final FTLHttpClient ftlClient;

  ShippingClient({required this.ftlClient});


  Future<currency.Money> getQuote(
    ShippingRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.post('/shipping/quote', request: request.toJson());
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return currency.Money.fromJson(body);
    } else {
      throw Exception('Failed to get getQuote response');
    }
  }

  Future<ShipOrderResponse> shipOrder(
    ShippingRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.post('/shipping/ship', request: request.toJson());
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return ShipOrderResponse.fromJson(body);
    } else {
      throw Exception('Failed to get shipOrder response');
    }
  }
}
