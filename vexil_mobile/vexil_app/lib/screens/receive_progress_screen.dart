import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/vexil_service.dart';
import '../utils/format.dart';
import '../providers/discovery_provider.dart';
import '../l10n/app_localizations.dart';

class ReceiveProgressScreen extends ConsumerStatefulWidget {
  final String taskId;
  final String saveDir;
  final int port;

  const ReceiveProgressScreen({
    super.key,
    required this.taskId,
    required this.saveDir,
    required this.port,
  });

  @override
  ConsumerState<ReceiveProgressScreen> createState() => _ReceiveProgressScreenState();
}

class _ReceiveProgressScreenState extends ConsumerState<ReceiveProgressScreen> {
  double _percent = 0;
  double _speedMBps = 0;
  int _sent = 0;
  int _total = 0;
  int _eta = 0;
  String _state = '等待连接中...';
  String _peerName = '';

  @override
  void initState() {
    super.initState();
    final service = ref.read(vexilServiceProvider);
    service.startForegroundService();
    _listenEvents();
  }

  void _listenEvents() {
    final loc = AppLocalizations.of(context);
    ref.read(vexilServiceProvider).events.listen((event) {
      if (event['taskId'] != widget.taskId) return;
      setState(() {
        switch (event['type']) {
          case 'progress':
            _percent = (event['percent'] as num).toDouble();
            _speedMBps = (event['speedMBps'] as num).toDouble();
            _sent = (event['sent'] as num).toInt();
            _total = (event['total'] as num).toInt();
            _eta = (event['eta'] as num).toInt();
            _state = event['state'] as String? ?? '';
            if (_total > 0) {
              ref.read(vexilServiceProvider).updateNotification(
                'Vexil ${loc.receiving} ${_percent.toStringAsFixed(0)}%',
                '${formatSize(_sent)} / ${formatSize(_total)}',
              );
            }
            break;
          case 'complete':
            _state = 'completed';
            _percent = 100;
            ref.read(vexilServiceProvider).updateNotification('Vexil', loc.receiveComplete);
            ref.read(vexilServiceProvider).stopForegroundService();
            break;
          case 'error':
            _state = 'failed';
            ref.read(vexilServiceProvider).updateNotification('Vexil', loc.failed);
            ref.read(vexilServiceProvider).stopForegroundService();
            break;
        }
      });
    });
  }

  @override
  Widget build(BuildContext context) {
    final loc = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(_state == 'completed' ? loc.receiveComplete : loc.receiving),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () {
            if (_state != 'completed') {
              ref.read(vexilServiceProvider).cancelTransfer(widget.taskId);
            }
            Navigator.pop(context);
          },
        ),
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('${loc.port}: ${widget.port}', style: const TextStyle(color: Colors.grey)),
              const SizedBox(height: 8),
              Text('${loc.saveTo}: ${widget.saveDir}', style: const TextStyle(color: Colors.grey, fontSize: 12)),
              const SizedBox(height: 24),
              LinearProgressIndicator(value: _total > 0 ? _percent / 100 : null),
              const SizedBox(height: 24),
              Text(
                _total > 0 ? '${_percent.toStringAsFixed(1)}%' : _state == 'waiting' ? loc.waitingConnection : loc.preparing,
                style: const TextStyle(fontSize: 32),
              ),
              if (_total > 0) ...[
                const SizedBox(height: 16),
                Text('${formatSize(_sent)} / ${formatSize(_total)}'),
                const SizedBox(height: 8),
                Text(formatSpeed(_speedMBps * 1024 * 1024)),
                if (_eta > 0) Text('剩余 ${formatETA(_eta)}'),
              ],
              const SizedBox(height: 16),
              if (_peerName.isNotEmpty) Text('${loc.from}: $_peerName'),
            ],
          ),
        ),
      ),
    );
  }
}