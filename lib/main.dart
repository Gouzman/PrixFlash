import 'package:flutter/material.dart';

import 'screens/home_screen.dart';
import 'theme.dart';

void main() {
  runApp(const PubprixApp());
}

class PubprixApp extends StatefulWidget {
  const PubprixApp({super.key});

  @override
  State<PubprixApp> createState() => _PubprixAppState();
}

class _PubprixAppState extends State<PubprixApp> {
  ThemeMode _mode = ThemeMode.dark;

  void _toggleTheme() {
    setState(() {
      _mode = _mode == ThemeMode.dark ? ThemeMode.light : ThemeMode.dark;
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Pubprix',
      debugShowCheckedModeBanner: false,
      themeMode: _mode,
      theme: pubprixLightTheme(),
      darkTheme: pubprixDarkTheme(),
      home: HomeScreen(onToggleTheme: _toggleTheme),
    );
  }
}
