import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:pubprix/main.dart';

void main() {
  testWidgets('Home screen shows the photo selection entry point', (WidgetTester tester) async {
    await tester.pumpWidget(const PubprixApp());

    expect(find.text('Sélectionnez vos photos'), findsOneWidget);
    expect(find.widgetWithText(ElevatedButton, 'Choisir des photos'), findsOneWidget);
  });
}
