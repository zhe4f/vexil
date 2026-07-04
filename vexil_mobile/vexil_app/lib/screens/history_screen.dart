import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/discovery_provider.dart';
import '../models/history_entry.dart';
import '../utils/format.dart';
import 'package:flutter/gestures.dart';
import '../l10n/app_localizations.dart';

class HistoryScreen extends ConsumerStatefulWidget {
  const HistoryScreen({super.key});

  @override
  ConsumerState<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends ConsumerState<HistoryScreen> {
  List<HistoryEntry> _entries = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    final service = ref.read(vexilServiceProvider);
    final entries = await service.getHistory();
    setState(() {
      _entries = entries;
      _loading = false;
    });
  }

  Future<void> _deleteItem(int index) async {
    final service = ref.read(vexilServiceProvider);
    await service.deleteHistory(index);
    _loadHistory();
  }

  Future<void> _clearAll() async {
    final loc = AppLocalizations.of(context);
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(loc.clearAllConfirm),
        content: Text(loc.clearAllConfirm),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(loc.cancel)),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: Text(loc.clearAll)),
        ],
      ),
    );
    if (confirm == true) {
      final service = ref.read(vexilServiceProvider);
      await service.clearHistory();
      _loadHistory();
    }
  }

  Future<void> _deleteRecord(int index) async {
    final service = ref.read(vexilServiceProvider);
    await service.deleteHistory(index + 1);
    _loadHistory();
  }

  Future<void> _deleteFile(HistoryEntry entry, int index) async {
    final loc = AppLocalizations.of(context);
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(loc.deleteConfirmTitle),
        content: Text(loc.deleteFileConfirm),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(loc.cancel)),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: Text(loc.deleteFile)),
        ],
      ),
    );
    if (confirm == true) {
      final service = ref.read(vexilServiceProvider);
      if (entry.savePath.isNotEmpty) {
        if (entry.direction == 'recv') {
          await service.deleteFile(entry.savePath, entry.fileNames);
        } else {
          await service.deleteFiles(entry.savePath);
        }
      }
      await service.deleteHistory(index + 1);
      _loadHistory();
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(loc.historyTitle),
        actions: _entries.isNotEmpty
            ? [
                IconButton(
                  icon: const Icon(Icons.delete_outline),
                  onPressed: _clearAll,
                ),
              ]
            : null,
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _entries.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.inbox, size: 64, color: Colors.grey),
                      const SizedBox(height: 16),
                      Text(loc.noHistory, style: const TextStyle(color: Colors.grey)),
                    ],
                  ),
                )
              : ListView.builder(
                  itemCount: _entries.length,
                  itemBuilder: (context, index) {
                    final entry = _entries[index];
                    final isSend = entry.direction == 'send';
                    return Dismissible(
                      key: Key('history_$index'),
                      direction: DismissDirection.endToStart,
                      background: Container(
                        color: Colors.red,
                        alignment: Alignment.centerRight,
                        padding: const EdgeInsets.only(right: 16),
                        child: const Icon(Icons.delete, color: Colors.white),
                      ),
                      onDismissed: (_) => _deleteItem(index + 1),
                      child: ListTile(
                        leading: CircleAvatar(
                          backgroundColor: entry.success
                              ? Colors.green.withAlpha(30)
                              : Colors.red.withAlpha(30),
                          child: Icon(
                            isSend ? Icons.arrow_upward : Icons.arrow_downward,
                            color: entry.success ? Colors.green : Colors.red,
                            size: 18,
                          ),
                        ),
                        title: RichText(
                          text: TextSpan(
                            style: DefaultTextStyle.of(context).style,
                            children: [
                              TextSpan(
                                text: entry.fileNames.length == 1
                                    ? entry.fileNames.first
                                    : '${entry.fileNames.first} ...',
                                style: const TextStyle(color: Colors.blue),
                                recognizer: TapGestureRecognizer()
                                  ..onTap = () {
                                    if (entry.savePath.isNotEmpty) {
                                      final service = ref.read(vexilServiceProvider);
                                      service.openFile(entry.savePath);
                                    }
                                  },
                              ),
                              TextSpan(
                                text: isSend
                                    ? ' → ${entry.peerName}'
                                    : ' ← ${entry.peerName}',
                              ),
                            ],
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        subtitle: Text(
                          '${entry.time}\n${formatSize(entry.size)} · ${entry.speedMBps.toStringAsFixed(1)} MB/s',
                        ),
                        trailing: Icon(
                          entry.success ? Icons.check_circle : Icons.error,
                          color: entry.success ? Colors.green : Colors.red,
                          size: 18,
                        ),
                        onTap: () => _showDetail(entry, index),
                      ),
                    );
                  },
                ),
    );
  }

  void _showDetail(HistoryEntry entry, int index) {
    final loc = AppLocalizations.of(context);
    showModalBottomSheet(
      context: context,
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              entry.direction == 'send' ? loc.detailSend : loc.detailReceive,
              style: Theme.of(ctx).textTheme.titleMedium,
            ),
            const SizedBox(height: 12),
            _detailRow(loc.time, entry.time),
            _detailRow(loc.peer, '${entry.peerName} (${entry.peer})'),
            _detailRow(loc.size, formatSize(entry.size)),
            if (entry.durationSec > 0) _detailRow(loc.duration, '${entry.durationSec.toStringAsFixed(1)}s'),
            if (entry.speedMBps > 0) _detailRow(loc.speed, '${entry.speedMBps.toStringAsFixed(1)} MB/s'),
            _detailRow(loc.fileCount, '${entry.files}'),
            if (entry.fileNames.isNotEmpty)
              _FileListRow(
                fileNames: entry.fileNames,
                savePath: entry.savePath,
                direction: entry.direction,
                onOpenFile: (path) {
                  final service = ref.read(vexilServiceProvider);
                  service.openFile(path);
                },
              ),
            if (entry.savePath.isNotEmpty) _detailRow(loc.savePath, entry.savePath),
            _detailRow(loc.status, entry.success ? loc.success : loc.failed),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                OutlinedButton(
                  onPressed: () {
                    Navigator.pop(ctx);
                    _deleteRecord(index);
                  },
                  child: Text(loc.deleteRecord),
                ),
                if (entry.direction == 'recv') ...[
                  const SizedBox(width: 12),
                  OutlinedButton(
                    onPressed: () {
                      Navigator.pop(ctx);
                      _deleteFile(entry, index);
                    },
                    child: Text(loc.deleteFile),
                  ),
                ],
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _detailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 64,
            child: Text(label, style: const TextStyle(color: Colors.grey, fontSize: 13)),
          ),
          Expanded(child: Text(value, style: const TextStyle(fontSize: 13))),
        ],
      ),
    );
  }
}

class _FileListRow extends StatefulWidget {
  final List<String> fileNames;
  final String savePath;
  final String direction;
  final void Function(String path) onOpenFile;

  const _FileListRow({
    required this.fileNames,
    required this.savePath,
    required this.direction,
    required this.onOpenFile,
  });

  @override
  State<_FileListRow> createState() => _FileListRowState();
}

class _FileListRowState extends State<_FileListRow> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final loc = AppLocalizations.of(context);
    final showFiles = _expanded ? widget.fileNames : widget.fileNames.take(2).toList();
    final hasMore = widget.fileNames.length > 2;

    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const SizedBox(
            width: 64,
            child: Text('', style: TextStyle(color: Colors.grey, fontSize: 13)),
          ),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ...showFiles.asMap().entries.map((e) {
                  final name = e.value;
                  final filePath = widget.direction == 'recv'
                      ? '${widget.savePath}/$name'
                      : widget.savePath;
                  return GestureDetector(
                    onTap: () => widget.onOpenFile(filePath),
                    child: Padding(
                      padding: const EdgeInsets.only(bottom: 2),
                      child: Text(
                        name,
                        style: const TextStyle(fontSize: 13, color: Colors.blue),
                      ),
                    ),
                  );
                }),
                if (hasMore)
                  GestureDetector(
                    onTap: () => setState(() => _expanded = !_expanded),
                    child: Text(
                      _expanded ? loc.collapse : loc.fileCountMore(widget.fileNames.length),
                      style: const TextStyle(fontSize: 13, color: Colors.blue),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}