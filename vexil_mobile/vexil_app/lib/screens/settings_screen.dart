import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:permission_handler/permission_handler.dart';
import '../l10n/app_localizations.dart';
import '../providers/locale_provider.dart';

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  int _numConns = 4;
  int _maxChunkMB = 16;
  bool _tlsEnabled = true;
  bool _loading = true;
  String _storageMode = 'external';
  String _currentLang = 'zh';

  @override
  void initState() {
    super.initState();
    _loadSettings();
    _loadLang();
  }

  Future<void> _loadSettings() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _numConns = prefs.getInt('num_conns') ?? 4;
      _maxChunkMB = prefs.getInt('max_chunk_mb') ?? 16;
      _tlsEnabled = prefs.getBool('tls_enabled') ?? true;
      _storageMode = prefs.getString('storage_mode') ?? 'external';
      _loading = false;
    });
  }

  Future<void> _loadLang() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _currentLang = prefs.getString('language') ?? 'zh';
    });
  }

  Future<void> _changeLanguage(String lang) async {
    await ref.read(localeProvider.notifier).setLocale(lang);
    setState(() => _currentLang = lang);
  }

  Future<void> _saveInt(String key, int value) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt(key, value);
  }

  Future<void> _saveBool(String key, bool value) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(key, value);
  }

  Future<void> _saveStorageMode(String mode) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('storage_mode', mode);
    setState(() => _storageMode = mode);
  }

  Future<void> _requestStoragePermission() async {
    final loc = AppLocalizations.of(context);
    var status = await Permission.storage.status;
    if (!status.isGranted) {
      status = await Permission.storage.request();
    }
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(status.isGranted ? loc.permissionGranted : loc.permissionDenied)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = AppLocalizations.of(context);
    if (_loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    return Scaffold(
      appBar: AppBar(title: Text(loc.settings)),
      body: ListView(
        children: [
          // 语言设置
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Text(loc.language, style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.grey)),
          ),
          RadioListTile<String>(
            title: Text(loc.chinese),
            value: 'zh',
            groupValue: _currentLang,
            onChanged: (v) => _changeLanguage(v!),
          ),
          RadioListTile<String>(
            title: Text(loc.english),
            value: 'en',
            groupValue: _currentLang,
            onChanged: (v) => _changeLanguage(v!),
          ),

          // 存储设置
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Text(loc.storageSettings, style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.grey)),
          ),
          RadioListTile<String>(
            title: Text(loc.publicDownloads),
            subtitle: Text(loc.publicDesc),
            value: 'external',
            groupValue: _storageMode,
            onChanged: (v) => _saveStorageMode(v!),
          ),
          RadioListTile<String>(
            title: Text(loc.appPrivate),
            subtitle: Text(loc.privateDesc),
            value: 'private',
            groupValue: _storageMode,
            onChanged: (v) => _saveStorageMode(v!),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: OutlinedButton(
              onPressed: _requestStoragePermission,
              child: Text(loc.requestPermission),
            ),
          ),

          const Divider(height: 32),
          // 传输设置
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
            child: Text(loc.transferSettings, style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.grey)),
          ),
          ListTile(
            title: Text(loc.concurrency),
            subtitle: Text('$_numConns'),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.remove),
                  onPressed: _numConns > 1
                      ? () { setState(() => _numConns--); _saveInt('num_conns', _numConns); }
                      : null,
                ),
                Text('$_numConns'),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: _numConns < 8
                      ? () { setState(() => _numConns++); _saveInt('num_conns', _numConns); }
                      : null,
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          ListTile(
            title: Text(loc.maxChunk),
            subtitle: Text('$_maxChunkMB MB'),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.remove),
                  onPressed: _maxChunkMB > 4
                      ? () { setState(() => _maxChunkMB -= 4); _saveInt('max_chunk_mb', _maxChunkMB); }
                      : null,
                ),
                Text('$_maxChunkMB'),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: _maxChunkMB < 64
                      ? () { setState(() => _maxChunkMB += 4); _saveInt('max_chunk_mb', _maxChunkMB); }
                      : null,
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          SwitchListTile(
            title: Text(loc.tlsEncryption),
            subtitle: Text(loc.tlsDesc),
            value: _tlsEnabled,
            onChanged: (v) { setState(() => _tlsEnabled = v); _saveBool('tls_enabled', _tlsEnabled); },
          ),
        ],
      ),
    );
  }
}