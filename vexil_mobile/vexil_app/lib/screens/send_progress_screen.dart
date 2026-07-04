import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/vexil_service.dart';
import '../utils/format.dart';
import '../providers/discovery_provider.dart';
import '../l10n/app_localizations.dart';

class SendProgressScreen extends ConsumerStatefulWidget {
  final String taskId;
  const SendProgressScreen({super.key, required this.taskId});

  @override
  ConsumerState<SendProgressScreen> createState() => _SendProgressScreenState();
}

class _SendProgressScreenState extends ConsumerState<SendProgressScreen> {
  double _percent = 0;
  double _speedMBps = 0;
  int _sent = 0;
  int _total = 0;
  int _eta = 0;
  String _state = 'preparing';

  @override
  void initState() {
    super.initState();
    _listenEvents();
  }

  void _listenEvents() {
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
            break;
          case 'complete':
            _state = 'completed';
            _percent = 100;
            break;
          case 'error':
            _state = 'failed';
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
        title: Text(_state == 'completed' ? loc.sendComplete : loc.sending),
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
              LinearProgressIndicator(value: _percent / 100),
              const SizedBox(height: 24),
              Text('${_percent.toStringAsFixed(1)}%', style: const TextStyle(fontSize: 32)),
              const SizedBox(height: 16),
              Text('${formatSize(_sent)} / ${formatSize(_total)}'),
              const SizedBox(height: 8),
              Text(formatSpeed(_speedMBps * 1024 * 1024)),
              if (_eta > 0) Text('剩余 ${formatETA(_eta)}'),
            ],
          ),
        ),
      ),
    );
  }
}