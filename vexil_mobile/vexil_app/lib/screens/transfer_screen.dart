import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:file_picker/file_picker.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../providers/discovery_provider.dart';
import '../widgets/device_card.dart';
import '../widgets/manual_connect_dialog.dart';
import '../services/vexil_service.dart';
import 'send_progress_screen.dart';
import 'receive_progress_screen.dart';
import 'package:flutter/services.dart';
import 'dart:io';
import 'package:path_provider/path_provider.dart';
import 'package:permission_handler/permission_handler.dart';
import '../l10n/app_localizations.dart';

class TransferScreen extends ConsumerStatefulWidget {
  const TransferScreen({super.key});

  @override
  ConsumerState<TransferScreen> createState() => _TransferScreenState();
}

class _TransferScreenState extends ConsumerState<TransferScreen> {
  final _nameCtrl = TextEditingController();
  String _localIP = '';

  @override
  void initState() {
    super.initState();
    _loadInfo();
  }

  Future<void> _loadInfo() async {
    final prefs = await SharedPreferences.getInstance();
    var name = prefs.getString('device_name') ?? '';

    if (name.isEmpty) {
      const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
      final random = DateTime.now().millisecondsSinceEpoch.toString();
      final suffix = random.substring(random.length - 4);
      name = 'Vexil-$suffix';
      await prefs.setString('device_name', name);
    }

    _nameCtrl.text = name;

    try {
      final result = await const MethodChannel('com.vexil/vexil').invokeMethod('getLocalIP');
      _localIP = result as String? ?? '';
    } catch (_) {}

    setState(() {});
  }

  Future<void> _saveName(String name) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('device_name', name);
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    super.dispose();
  }

  Future<void> _startReceive(String dir) async {
    final port = 54321;
    final service = ref.read(vexilServiceProvider);
    final taskId = await service.startReceive(port, dir);
    if (context.mounted) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => ReceiveProgressScreen(
            taskId: taskId,
            saveDir: dir,
            port: port,
          ),
        ),
      );
    }
  }

  static int get _androidSdkVersion {
    if (!Platform.isAndroid) return 0;
    try {
      return int.tryParse(Platform.version.split(' ')[0]) ?? 0;
    } catch (_) {
      return 0;
    }
  }

  Future<String?> _getSaveDir() async {
    final prefs = await SharedPreferences.getInstance();
    final storageMode = prefs.getString('storage_mode') ?? 'external';
    final loc = AppLocalizations.of(context);

    if (storageMode == 'private') {
      final appDir = await getApplicationDocumentsDirectory();
      final dir = '${appDir.path}/Vexil';
      await Directory(dir).create(recursive: true);
      return dir;
    }

    final savedPath = prefs.getString('saved_public_path');
    if (savedPath != null) {
      final canWrite = await const MethodChannel('com.vexil/vexil')
          .invokeMethod('checkWritePermission', {'path': savedPath});
      if (canWrite == true) {
        await Directory(savedPath).create(recursive: true);
        return savedPath;
      }
    }

    if (_androidSdkVersion >= 30) {
      if (!context.mounted) return null;
      final dir = await FilePicker.platform.getDirectoryPath(
        dialogTitle: loc.selectSaveDir,
      );
      if (dir == null) return null;
      final canWrite = await const MethodChannel('com.vexil/vexil')
          .invokeMethod('checkWritePermission', {'path': dir});
      if (canWrite == true) {
        await prefs.setString('saved_public_path', dir);
        await Directory(dir).create(recursive: true);
        return dir;
      }
    } else {
      var status = await Permission.storage.status;
      if (!status.isGranted) {
        status = await Permission.storage.request();
      }
      if (status.isGranted) {
        final dir = '/storage/emulated/0/Download/Vexil';
        await Directory(dir).create(recursive: true);
        final canWrite = await const MethodChannel('com.vexil/vexil')
            .invokeMethod('checkWritePermission', {'path': dir});
        if (canWrite == true) {
          await prefs.setString('saved_public_path', dir);
          return dir;
        }
      }
    }

    if (context.mounted) {
      await showDialog(
        context: context,
        builder: (ctx) => AlertDialog(
          title: Text(loc.cannotUsePublic),
          content: Text(loc.fallbackToPrivate),
          actions: [
            FilledButton(
              onPressed: () => Navigator.pop(ctx),
              child: Text(loc.ok),
            ),
          ],
        ),
      );
    }
    final appDir = await getApplicationDocumentsDirectory();
    final dir = '${appDir.path}/Vexil';
    await Directory(dir).create(recursive: true);
    return dir;
  }

  @override
  Widget build(BuildContext context) {
    final devicesAsync = ref.watch(discoveryStateProvider);
    final loc = AppLocalizations.of(context);

    return Scaffold(
      appBar: AppBar(title: Text(loc.appTitle)),
      body: Column(
        children: [
          Container(
            margin: const EdgeInsets.all(12),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _nameCtrl,
                    decoration: InputDecoration(
                      labelText: loc.deviceName,
                      border: const OutlineInputBorder(),
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    ),
                    onChanged: _saveName,
                  ),
                ),
                const SizedBox(width: 12),
                if (_localIP.isNotEmpty)
                  Text('IP: $_localIP', style: const TextStyle(color: Colors.grey)),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Row(
              children: [
                Expanded(
                  child: Text(loc.nearbyDevices, style: Theme.of(context).textTheme.titleMedium),
                ),
                FilledButton.icon(
                  onPressed: () => ref.read(discoveryStateProvider.notifier).scan(),
                  icon: const Icon(Icons.refresh, size: 18),
                  label: Text(loc.scan),
                ),
              ],
            ),
          ),
          const SizedBox(height: 8),
          Expanded(
            child: devicesAsync.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (err, _) => Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.wifi_off, size: 48, color: Colors.grey),
                    const SizedBox(height: 8),
                    Text('${loc.scanFailed}$err'),
                  ],
                ),
              ),
              data: (devices) {
                if (devices.isEmpty) {
                  return Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.devices_other, size: 48, color: Colors.grey),
                        const SizedBox(height: 8),
                        Text(loc.noDeviceFound),
                      ],
                    ),
                  );
                }
                return ListView.builder(
                  itemCount: devices.length,
                  itemBuilder: (context, index) {
                    return DeviceCard(
                      device: devices[index],
                      onTap: () async {
                        final device = devices[index];
                        final result = await FilePicker.platform.pickFiles(allowMultiple: true);
                        if (result == null || result.files.isEmpty || !context.mounted) return;
                        final paths = result.files.map((f) => f.path ?? '').where((p) => p.isNotEmpty).toList();
                        if (paths.isEmpty) return;
                        final service = ref.read(vexilServiceProvider);
                        final taskId = await service.sendFiles(device.ip, device.port, paths, device.name);
                        if (context.mounted) {
                          Navigator.push(
                            context,
                            MaterialPageRoute(builder: (_) => SendProgressScreen(taskId: taskId)),
                          );
                        }
                      },
                    );
                  },
                );
              },
            ),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(   // 手动发送变为透明框
                      onPressed: () async {
                        final result = await showDialog<Map<String, dynamic>>(
                          context: context,
                          builder: (ctx) => const ManualConnectDialog(),
                        );
                        if (result != null && context.mounted) {
                          final ip = result['ip'] as String;
                          final port = result['port'] as int;
                          final r = await FilePicker.platform.pickFiles(allowMultiple: true);
                          if (r == null || r.files.isEmpty || !context.mounted) return;
                          final paths = r.files.map((f) => f.path ?? '').where((p) => p.isNotEmpty).toList();
                          if (paths.isEmpty) return;
                          final service = ref.read(vexilServiceProvider);
                          final taskId = await service.sendFiles(ip, port, paths, ip);
                          if (context.mounted) {
                            Navigator.push(
                              context,
                              MaterialPageRoute(builder: (_) => SendProgressScreen(taskId: taskId)),
                            );
                          }
                        }
                      },
                      icon: const Icon(Icons.send),
                      label: Text(loc.manualSend),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: FilledButton.icon(    // 开始接收变为填充框
                      onPressed: () async {
                        try {
                          final dir = await _getSaveDir();
                          if (dir != null) {
                            _startReceive(dir);
                          }
                        } catch (e) {
                          if (context.mounted) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(content: Text('${loc.error}: $e')),
                            );
                          }
                        }
                      },
                      icon: const Icon(Icons.download),
                      label: Text(loc.startReceive),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
