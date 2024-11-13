// ignore_for_file: unused_import
library builtin;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';


class Empty{

  Empty();

  Map<String, dynamic> toJson() {
    return {
    };
  }

  factory Empty.fromJson(Map<String, dynamic> map) {
    return Empty(
      
    );
  }
}

class HttpRequest<Body>{
  String method;
  String path;
  Map<String, String> pathParameters;
  Map<String, List<String>> query;
  Map<String, List<String>> headers;
  Body body;

  HttpRequest({  required this.method,  required this.path,  required this.pathParameters,  required this.query,  required this.headers,  required this.body,  });

  Map<String, dynamic> toJson() {
    return {
      'method': ((dynamic v) => v)(method),
      'path': ((dynamic v) => v)(path),
      'pathParameters': ((dynamic v) => v.map((k, v) => MapEntry(k, v)).cast<String, String>())(pathParameters),
      'query': ((dynamic v) => v.map((k, v) => MapEntry(k, v.map((v) => v).cast<String>().toList())).cast<String, List<String>>())(query),
      'headers': ((dynamic v) => v.map((k, v) => MapEntry(k, v.map((v) => v).cast<String>().toList())).cast<String, List<String>>())(headers),
      'body': ((dynamic v) => v.toJson())(body),
    };
  }

  factory HttpRequest.fromJson(Map<String, dynamic> map, Body Function(Map<String, dynamic>) bodyJsonFn) {
    return HttpRequest(
      method: ((dynamic v) => v)(map['method']), path: ((dynamic v) => v)(map['path']), pathParameters: ((dynamic v) => v.map((k, v) => MapEntry(k, v)).cast<String, String>())(map['pathParameters']), query: ((dynamic v) => v.map((k, v) => MapEntry(k, v.map((v) => v).cast<String>().toList())).cast<String, List<String>>())(map['query']), headers: ((dynamic v) => v.map((k, v) => MapEntry(k, v.map((v) => v).cast<String>().toList())).cast<String, List<String>>())(map['headers']), body: bodyJsonFn(map['body']), 
    );
  }
}

class HttpResponse<Body, Error>{
  int status;
  Map<String, List<String>> headers;
  Body? body;
  Error? error;

  HttpResponse({  required this.status,  required this.headers,  this.body,  this.error,  });

  Map<String, dynamic> toJson() {
    return {
      'status': ((dynamic v) => v)(status),
      'headers': ((dynamic v) => v.map((k, v) => MapEntry(k, v.map((v) => v).cast<String>().toList())).cast<String, List<String>>())(headers),
      'body': ((dynamic v) => v)(body),
      'error': ((dynamic v) => v)(error),
    };
  }

  factory HttpResponse.fromJson(Map<String, dynamic> map, Body Function(Map<String, dynamic>) bodyJsonFn, Error Function(Map<String, dynamic>) errorJsonFn) {
    return HttpResponse(
      status: ((dynamic v) => v)(map['status']), headers: ((dynamic v) => v.map((k, v) => MapEntry(k, v.map((v) => v).cast<String>().toList())).cast<String, List<String>>())(map['headers']), body: ((dynamic v) => v)(map['body']), error: ((dynamic v) => v)(map['error']), 
    );
  }
}


class BuiltinClient {
  final FTLHttpClient ftlClient;

  BuiltinClient({required this.ftlClient});

}
