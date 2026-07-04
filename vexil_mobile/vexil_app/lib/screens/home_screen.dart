import 'package:flutter/material.dart';
import 'transfer_screen.dart';
import 'history_screen.dart';
import 'settings_screen.dart';
import '../l10n/app_localizations.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _currentIndex = 0;

  final _screens = const [
    TransferScreen(),
    HistoryScreen(),
    SettingsScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    final loc = AppLocalizations.of(context);
    return Scaffold(
      body: _screens[_currentIndex],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (i) => setState(() => _currentIndex = i),
        destinations: [
          NavigationDestination(
            icon: const Icon(Icons.swap_horiz),
            label: loc.tabTransfer,
          ),
          NavigationDestination(
            icon: const Icon(Icons.history),
            label: loc.tabHistory,
          ),
          NavigationDestination(
            icon: const Icon(Icons.settings),
            label: loc.tabSettings,
          ),
        ],
      ),
    );
  }
}