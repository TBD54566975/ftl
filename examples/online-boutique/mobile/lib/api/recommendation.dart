// ignore_for_file: unused_import
library recommendation;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
import 'builtin.dart' as builtin;
import 'productcatalog.dart' as productcatalog;


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

class ListRequest{
  String userId;
  List<String> userProductIDs;

  ListRequest({  required this.userId,  required this.userProductIDs,  });

  Map<String, dynamic> toJson() {
    return {
      'userId': ((dynamic v) => v)(userId),
      'userProductIDs': ((dynamic v) => v.map((v) => v).cast<String>().toList())(userProductIDs),
    };
  }

  factory ListRequest.fromJson(Map<String, dynamic> map) {
    return ListRequest(
      userId: ((dynamic v) => v)(map['userId']), userProductIDs: ((dynamic v) => v.map((v) => v).cast<String>().toList())(map['userProductIDs']), 
    );
  }
}

class ListResponse{
  List<String> productIDs;

  ListResponse({  required this.productIDs,  });

  Map<String, dynamic> toJson() {
    return {
      'productIDs': ((dynamic v) => v.map((v) => v).cast<String>().toList())(productIDs),
    };
  }

  factory ListResponse.fromJson(Map<String, dynamic> map) {
    return ListResponse(
      productIDs: ((dynamic v) => v.map((v) => v).cast<String>().toList())(map['productIDs']), 
    );
  }
}


class RecommendationClient {
  final FTLHttpClient ftlClient;

  RecommendationClient({required this.ftlClient});


  Future<ListResponse> list(
    ListRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.get(
      '/recommendation', 
      requestJson: json.encode(request.toJson()),
      headers: headers,
    );
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return ListResponse.fromJson(body);
    } else {
      throw Exception('Failed to get list response');
    }
  }

}
