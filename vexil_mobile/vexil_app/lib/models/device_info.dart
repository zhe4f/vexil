class DeviceInfo {
  final String name;
  final String ip;
  final int port;
  final String source;

  const DeviceInfo({
    required this.name,
    required this.ip,
    required this.port,
    this.source = 'udp',
  });

  factory DeviceInfo.fromJson(Map<String, dynamic> json) {
    return DeviceInfo(
      name: json['name'] as String? ?? '',
      ip: json['ip'] as String? ?? '',
      port: json['port'] as int? ?? 0,
      source: json['source'] as String? ?? 'udp',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'ip': ip,
      'port': port,
      'source': source,
    };
  }
}