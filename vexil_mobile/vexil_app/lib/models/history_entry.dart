class HistoryEntry {
  final String time;
  final String direction; // "send" / "recv"
  final String peer;
  final String peerName;
  final int files;
  final List<String> fileNames;
  final int size;
  final double durationSec;
  final double speedMBps;
  final bool success;
  final String savePath;

  const HistoryEntry({
    required this.time,
    required this.direction,
    required this.peer,
    this.peerName = '',
    required this.files,
    required this.fileNames,
    required this.size,
    required this.durationSec,
    required this.speedMBps,
    required this.success,
    this.savePath = '',
  });

  factory HistoryEntry.fromJson(Map<String, dynamic> json) {
    final fileNamesRaw = json['file_names'];
    final fileNames = fileNamesRaw != null ? List<String>.from(fileNamesRaw) : <String>[];
    
    return HistoryEntry(
      time: json['time'] as String? ?? '',
      direction: json['direction'] as String? ?? '',
      peer: json['peer'] as String? ?? '',
      peerName: json['peer_name'] as String? ?? '',
      files: json['files'] as int? ?? 0,
      fileNames: fileNames,
      size: json['size'] as int? ?? 0,
      durationSec: (json['duration_sec'] as num?)?.toDouble() ?? 0,
      speedMBps: (json['speed_mbps'] as num?)?.toDouble() ?? 0,
      success: json['success'] as bool? ?? false,
      savePath: json['save_path'] as String? ?? '',
    );
  }
}