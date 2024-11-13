// ignore_for_file: unused_import
library payment;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
import 'builtin.dart' as builtin;
import 'currency.dart' as currency;


class ChargeRequest{
  currency.Money amount;
  CreditCardInfo creditCard;

  ChargeRequest({  required this.amount,  required this.creditCard,  });

  Map<String, dynamic> toJson() {
    return {
      'amount': ((dynamic v) => v.toJson())(amount),
      'creditCard': ((dynamic v) => v.toJson())(creditCard),
    };
  }

  factory ChargeRequest.fromJson(Map<String, dynamic> map) {
    return ChargeRequest(
      amount: ((dynamic v) => currency.Money.fromJson(v))(map['amount']), creditCard: ((dynamic v) => CreditCardInfo.fromJson(v))(map['creditCard']), 
    );
  }
}

class ChargeResponse{
  String transactionId;

  ChargeResponse({  required this.transactionId,  });

  Map<String, dynamic> toJson() {
    return {
      'transactionId': ((dynamic v) => v)(transactionId),
    };
  }

  factory ChargeResponse.fromJson(Map<String, dynamic> map) {
    return ChargeResponse(
      transactionId: ((dynamic v) => v)(map['transactionId']), 
    );
  }
}

class CreditCardInfo{
  String number;
  int cvv;
  int expirationYear;
  int expirationMonth;

  CreditCardInfo({  required this.number,  required this.cvv,  required this.expirationYear,  required this.expirationMonth,  });

  Map<String, dynamic> toJson() {
    return {
      'number': ((dynamic v) => v)(number),
      'cvv': ((dynamic v) => v)(cvv),
      'expirationYear': ((dynamic v) => v)(expirationYear),
      'expirationMonth': ((dynamic v) => v)(expirationMonth),
    };
  }

  factory CreditCardInfo.fromJson(Map<String, dynamic> map) {
    return CreditCardInfo(
      number: ((dynamic v) => v)(map['number']), cvv: ((dynamic v) => v)(map['cvv']), expirationYear: ((dynamic v) => v)(map['expirationYear']), expirationMonth: ((dynamic v) => v)(map['expirationMonth']), 
    );
  }
}

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


class PaymentClient {
  final FTLHttpClient ftlClient;

  PaymentClient({required this.ftlClient});


  Future<ChargeResponse> charge(
    ChargeRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.post('/payment/charge', request: request.toJson());
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return ChargeResponse.fromJson(body);
    } else {
      throw Exception('Failed to get charge response');
    }
  }
}
