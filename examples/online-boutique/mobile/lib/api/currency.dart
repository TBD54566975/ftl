// ignore_for_file: unused_import
library currency;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
import 'builtin.dart' as builtin;


class ConvertRequest{
  Money from;
  String toCode;

  ConvertRequest({  required this.from,  required this.toCode,  });

  Map<String, dynamic> toJson() {
    return {
      'from': ((dynamic v) => v.toJson())(from),
      'toCode': ((dynamic v) => v)(toCode),
    };
  }

  factory ConvertRequest.fromJson(Map<String, dynamic> map) {
    return ConvertRequest(
      from: ((dynamic v) => Money.fromJson(v))(map['from']), toCode: ((dynamic v) => v)(map['toCode']), 
    );
  }
}

class GetSupportedCurrenciesRequest{

  GetSupportedCurrenciesRequest();

  Map<String, dynamic> toJson() {
    return {
    };
  }

  factory GetSupportedCurrenciesRequest.fromJson(Map<String, dynamic> map) {
    return GetSupportedCurrenciesRequest(
      
    );
  }
}

class GetSupportedCurrenciesResponse{
  List<String> currencyCodes;

  GetSupportedCurrenciesResponse({  required this.currencyCodes,  });

  Map<String, dynamic> toJson() {
    return {
      'currencyCodes': ((dynamic v) => v.map((v) => v).cast<String>().toList())(currencyCodes),
    };
  }

  factory GetSupportedCurrenciesResponse.fromJson(Map<String, dynamic> map) {
    return GetSupportedCurrenciesResponse(
      currencyCodes: ((dynamic v) => v.map((v) => v).cast<String>().toList())(map['currencyCodes']), 
    );
  }
}

class Money{
  String currencyCode;
  int units;
  int nanos;

  Money({  required this.currencyCode,  required this.units,  required this.nanos,  });

  Map<String, dynamic> toJson() {
    return {
      'currencyCode': ((dynamic v) => v)(currencyCode),
      'units': ((dynamic v) => v)(units),
      'nanos': ((dynamic v) => v)(nanos),
    };
  }

  factory Money.fromJson(Map<String, dynamic> map) {
    return Money(
      currencyCode: ((dynamic v) => v)(map['currencyCode']), units: ((dynamic v) => v)(map['units']), nanos: ((dynamic v) => v)(map['nanos']), 
    );
  }
}


class CurrencyClient {
  final FTLHttpClient ftlClient;

  CurrencyClient({required this.ftlClient});


  Future<Money> convert(
    ConvertRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.post('/currency/convert', request: request.toJson());
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return Money.fromJson(body);
    } else {
      throw Exception('Failed to get convert response');
    }
  }

  Future<GetSupportedCurrenciesResponse> getSupportedCurrencies(
    GetSupportedCurrenciesRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.get(
      '/currency/supported', 
      requestJson: json.encode(request.toJson()),
      headers: headers,
    );
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return GetSupportedCurrenciesResponse.fromJson(body);
    } else {
      throw Exception('Failed to get getSupportedCurrencies response');
    }
  }
}
