import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:online_boutique/api/productcatalog.dart';

import 'package:online_boutique/main.dart';
import 'package:online_boutique/utils/api_providers.dart';

import 'helpers/mocks.dart';
import 'helpers/widget_test_helpers.dart';

void main() {
  late MockProductcatalogClient productcatalogClient;

  setUpAll(() {
    registerFallbackValue(ListRequest());
  });

  setUp(() {
    productcatalogClient = MockProductcatalogClient();
  });

  testWidgets('Renders app', (WidgetTester tester) async {
    when(() => productcatalogClient.list(any())).thenAnswer(
      (_) async => ListResponse(products: []),
    );

    await tester.pumpWidget(
      WidgetTestHelpers.testableWidget(
        child: const App(),
        overrides: [
          productsCatalogProvider.overrideWithValue(productcatalogClient)
        ],
      ),
    );
  });
}
