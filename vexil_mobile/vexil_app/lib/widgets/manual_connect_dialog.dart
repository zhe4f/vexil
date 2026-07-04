import 'package:flutter/material.dart';
import '../l10n/app_localizations.dart';

class ManualConnectDialog extends StatefulWidget {
  const ManualConnectDialog({super.key});

  @override
  State<ManualConnectDialog> createState() => _ManualConnectDialogState();
}

class _ManualConnectDialogState extends State<ManualConnectDialog> {
  final ipCtrl = TextEditingController();
  final portCtrl = TextEditingController(text: '54321');

  @override
  void dispose() {
    ipCtrl.dispose();
    portCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final loc = AppLocalizations.of(context);
    return AlertDialog(
      title: Text(loc.manualConnectTitle),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: ipCtrl,
            decoration: InputDecoration(
              labelText: loc.ipAddress,
              hintText: loc.ipHint,
            ),
            keyboardType: TextInputType.number,
          ),
          const SizedBox(height: 12),
          TextField(
            controller: portCtrl,
            decoration: InputDecoration(labelText: loc.port),
            keyboardType: TextInputType.number,
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(loc.cancel),
        ),
        FilledButton(
          onPressed: () {
            final ip = ipCtrl.text.trim();
            final port = int.tryParse(portCtrl.text.trim()) ?? 0;
            Navigator.pop(context, {'ip': ip, 'port': port});
          },
          child: Text(loc.next),
        ),
      ],
    );
  }
}