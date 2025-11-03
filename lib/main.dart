import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'core/theme/app_theme.dart';
import 'presentation/pages/home_page.dart';
import 'presentation/pages/editor_page.dart';

void main() {
  runApp(const ProviderScope(child: PrixFlashApp()));
}

class PrixFlashApp extends StatelessWidget {
  const PrixFlashApp({super.key});

  @override
  Widget build(BuildContext context) {
    final router = GoRouter(
      routes: [
        GoRoute(path: '/', builder: (context, state) => const HomePage()),
        GoRoute(
          path: '/editor',
          builder: (context, state) => const EditorPage(),
        ),
      ],
    );

    return MaterialApp.router(
      title: 'PrixFlash',
      theme: AppTheme.lightTheme,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
