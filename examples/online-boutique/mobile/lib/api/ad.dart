// ignore_for_file: unused_import
library ad;

import 'dart:convert';
import 'dart:typed_data';
import 'ftl_client.dart';
import 'builtin.dart' as builtin;


class Ad{
  String redirectUrl;
  String text;

  Ad({  required this.redirectUrl,  required this.text,  });

  Map<String, dynamic> toJson() {
    return {
      'redirectUrl': ((dynamic v) => v)(redirectUrl),
      'text': ((dynamic v) => v)(text),
    };
  }

  factory Ad.fromJson(Map<String, dynamic> map) {
    return Ad(
      redirectUrl: ((dynamic v) => v)(map['redirectUrl']), text: ((dynamic v) => v)(map['text']), 
    );
  }
}

class AdRequest{
  List<String> contextKeys;

  AdRequest({  required this.contextKeys,  });

  Map<String, dynamic> toJson() {
    return {
      'contextKeys': ((dynamic v) => v.map((v) => v).cast<String>().toList())(contextKeys),
    };
  }

  factory AdRequest.fromJson(Map<String, dynamic> map) {
    return AdRequest(
      contextKeys: ((dynamic v) => v.map((v) => v).cast<String>().toList())(map['contextKeys']), 
    );
  }
}

class AdResponse{
  String name;
  List<Ad> ads;

  AdResponse({  required this.name,  required this.ads,  });

  Map<String, dynamic> toJson() {
    return {
      'name': ((dynamic v) => v)(name),
      'ads': ((dynamic v) => v.map((v) => v.toJson()).cast<Ad>().toList())(ads),
    };
  }

  factory AdResponse.fromJson(Map<String, dynamic> map) {
    return AdResponse(
      name: ((dynamic v) => v)(map['name']), ads: ((dynamic v) => v.map((v) => Ad.fromJson(v)).cast<Ad>().toList())(map['ads']), 
    );
  }
}


class AdClient {
  final FTLHttpClient ftlClient;

  AdClient({required this.ftlClient});


  Future<AdResponse> get(
    AdRequest request, { 
    Map<String, String>? headers,
  }) async {
    final response = await ftlClient.get(
      '/ad', 
      requestJson: json.encode(request.toJson()),
      headers: headers,
    );
    if (response.statusCode == 200) {
      final body = json.decode(utf8.decode(response.bodyBytes));
      return AdResponse.fromJson(body);
    } else {
      throw Exception('Failed to get get response');
    }
  }
}
