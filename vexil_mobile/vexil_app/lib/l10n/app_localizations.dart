import 'package:flutter/material.dart';

class AppLocalizations {
  final Locale locale;

  AppLocalizations(this.locale);

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const _localizedValues = <String, Map<String, String>>{
    // 通用
    'appTitle': {'en': 'Vexil', 'zh': 'Vexil'},
    'cancel': {'en': 'Cancel', 'zh': '取消'},
    'confirm': {'en': 'Confirm', 'zh': '确定'},
    'close': {'en': 'Close', 'zh': '关闭'},
    'ok': {'en': 'OK', 'zh': '知道了'},

    // 底部导航
    'tabTransfer': {'en': 'Transfer', 'zh': '传输'},
    'tabHistory': {'en': 'History', 'zh': '历史'},
    'tabSettings': {'en': 'Settings', 'zh': '设置'},

    // 传输主页
    'deviceName': {'en': 'Device Name', 'zh': '设备名称'},
    'nearbyDevices': {'en': 'Nearby Devices', 'zh': '附近设备'},
    'scan': {'en': 'Scan', 'zh': '扫描'},
    'startReceive': {'en': 'Receive', 'zh': '开始接收'},
    'manualSend': {'en': 'Send', 'zh': '手动发送'},
    'noDeviceFound': {'en': 'No devices found\nMake sure you are on the same LAN', 'zh': '未发现设备\n请确保在同一局域网'},
    'scanFailed': {'en': 'Scan failed: ', 'zh': '扫描失败: '},

    // 发送/接收进度
    'sending': {'en': 'Sending', 'zh': '发送中'},
    'receiving': {'en': 'Receiving', 'zh': '接收中'},
    'sendComplete': {'en': 'Send Complete', 'zh': '发送完成'},
    'receiveComplete': {'en': 'Receive Complete', 'zh': '接收完成'},
    'preparing': {'en': 'Preparing...', 'zh': '准备中...'},
    'connecting': {'en': 'Connecting...', 'zh': '连接中...'},
    'running': {'en': 'Transferring...', 'zh': '传输中...'},
    'finalizing': {'en': 'Finalizing...', 'zh': '校验中...'},
    'waitingConnection': {'en': 'Waiting for connection...', 'zh': '等待连接中...'},
    'saveTo': {'en': 'Save to', 'zh': '保存到'},
    'from': {'en': 'From', 'zh': '来自'},
    'remaining': {'en': 'remaining ', 'zh': '剩余 '},

    // 手动连接弹窗
    'manualConnectTitle': {'en': 'Enter Address', 'zh': '手动输入目标'},
    'ipAddress': {'en': 'IP Address', 'zh': 'IP 地址'},
    'ipHint': {'en': 'e.g. 192.168.1.100', 'zh': '例如 192.168.1.100'},
    'port': {'en': 'Port', 'zh': '端口'},
    'next': {'en': 'Next', 'zh': '下一步'},

    // 历史页面
    'historyTitle': {'en': 'Transfer History', 'zh': '传输历史'},
    'noHistory': {'en': 'No transfer records', 'zh': '暂无传输记录'},
    'clearAll': {'en': 'Clear All', 'zh': '清空全部记录'},
    'deleteRecord': {'en': 'Delete Record', 'zh': '删除记录'},
    'deleteFile': {'en': 'Delete File', 'zh': '删除文件'},
    'deleteConfirmTitle': {'en': 'Confirm Delete', 'zh': '确认删除'},
    'deleteFileConfirm': {'en': 'Delete this file?', 'zh': '确定要删除该文件吗？'},
    'clearAllConfirm': {'en': 'Clear all records?', 'zh': '确定要清空所有传输记录吗？'},

    // 历史详情
    'detailSend': {'en': 'Sent', 'zh': '发送'},
    'detailReceive': {'en': 'Received', 'zh': '接收'},
    'time': {'en': 'Time', 'zh': '时间'},
    'peer': {'en': 'Peer', 'zh': '对方'},
    'size': {'en': 'Size', 'zh': '大小'},
    'duration': {'en': 'Duration', 'zh': '耗时'},
    'speed': {'en': 'Speed', 'zh': '速度'},
    'fileCount': {'en': 'Files', 'zh': '文件数'},
    'files': {'en': 'Files', 'zh': '文件'},
    'savePath': {'en': 'Save Path', 'zh': '保存路径'},
    'status': {'en': 'Status', 'zh': '状态'},
    'success': {'en': 'Success', 'zh': '成功'},
    'failed': {'en': 'Failed', 'zh': '失败'},
    'expand': {'en': 'Show all', 'zh': '展开'},
    'collapse': {'en': 'Collapse', 'zh': '收起'},
    'fileCountMore': {'en': '... {count} files', 'zh': '... 共{count}个文件'},

    // 设置页面
    'settings': {'en': 'Settings', 'zh': '设置'},
    'language': {'en': 'Language', 'zh': '语言'},
    'chinese': {'en': 'Chinese', 'zh': '中文'},
    'english': {'en': 'English', 'zh': 'English'},
    'storageSettings': {'en': 'Storage Settings', 'zh': '存储设置'},
    'publicDownloads': {'en': 'Public Downloads', 'zh': '公共下载目录'},
    'publicDesc': {'en': 'Save to public directory, files kept after uninstall', 'zh': '保存到公共目录，删 App 文件不丢失'},
    'appPrivate': {'en': 'App Private', 'zh': '应用私有目录'},
    'privateDesc': {'en': 'Save to app internal, files deleted with uninstall', 'zh': '保存到应用内部，删 App 文件同步删除'},
    'requestPermission': {'en': 'Request Permission', 'zh': '申请存储权限'},
    'transferSettings': {'en': 'Transfer Settings', 'zh': '传输设置'},
    'concurrency': {'en': 'Concurrent Connections', 'zh': '并发连接数'},
    'maxChunk': {'en': 'Max Chunk Size', 'zh': '最大块大小'},
    'securitySettings': {'en': 'Security Settings', 'zh': '安全设置'},
    'tlsEncryption': {'en': 'TLS Encryption', 'zh': 'TLS 加密'},
    'tlsDesc': {'en': 'End-to-end encryption during transfer', 'zh': '启用后传输内容端到端加密'},
    'permissionGranted': {'en': 'Permission granted', 'zh': '存储权限已授予'},
    'permissionDenied': {'en': 'Permission denied', 'zh': '权限未授予'},

    // 目录选择 / 存储提示
    'selectSaveDir': {'en': 'Select save directory', 'zh': '选择保存目录'},
    'selectSaveDirDesc': {'en': 'e.g. Download/Vexil', 'zh': '例如 Download/Vexil'},
    'cannotUsePublic': {'en': 'Cannot use public directory', 'zh': '无法使用公共目录'},
    'fallbackToPrivate': {'en': 'Files will be saved to app private directory.\nUninstalling the app will delete these files.', 'zh': '文件将保存到应用私有目录。\n删除应用会同时删除传输的文件。'},
    'noWritePermission': {'en': 'No write permission', 'zh': '目录无写入权限'},
    'chooseAction': {'en': 'Choose action', 'zh': '请选择处理方式'},
    'useInternalStorage': {'en': 'Use internal storage', 'zh': '使用内部存储'},
    'requestPermissionAndRetry': {'en': 'Request permission and retry', 'zh': '申请权限后重试'},
    'switchedToInternal': {'en': 'Switched to internal storage', 'zh': '已切换到应用内部存储'},
    'error': {'en': 'Error', 'zh': '错误'},
  };

  // 便捷 getter
  String get appTitle => _localizedValues['appTitle']![locale.languageCode] ?? '';
  String get cancel => _localizedValues['cancel']![locale.languageCode] ?? '';
  String get confirm => _localizedValues['confirm']![locale.languageCode] ?? '';
  String get close => _localizedValues['close']![locale.languageCode] ?? '';
  String get ok => _localizedValues['ok']![locale.languageCode] ?? '';

  String get tabTransfer => _localizedValues['tabTransfer']![locale.languageCode] ?? '';
  String get tabHistory => _localizedValues['tabHistory']![locale.languageCode] ?? '';
  String get tabSettings => _localizedValues['tabSettings']![locale.languageCode] ?? '';

  String get deviceName => _localizedValues['deviceName']![locale.languageCode] ?? '';
  String get nearbyDevices => _localizedValues['nearbyDevices']![locale.languageCode] ?? '';
  String get scan => _localizedValues['scan']![locale.languageCode] ?? '';
  String get startReceive => _localizedValues['startReceive']![locale.languageCode] ?? '';
  String get manualSend => _localizedValues['manualSend']![locale.languageCode] ?? '';
  String get noDeviceFound => _localizedValues['noDeviceFound']![locale.languageCode] ?? '';
  String get scanFailed => _localizedValues['scanFailed']![locale.languageCode] ?? '';

  String get sending => _localizedValues['sending']![locale.languageCode] ?? '';
  String get receiving => _localizedValues['receiving']![locale.languageCode] ?? '';
  String get sendComplete => _localizedValues['sendComplete']![locale.languageCode] ?? '';
  String get receiveComplete => _localizedValues['receiveComplete']![locale.languageCode] ?? '';
  String get preparing => _localizedValues['preparing']![locale.languageCode] ?? '';
  String get connecting => _localizedValues['connecting']![locale.languageCode] ?? '';
  String get running => _localizedValues['running']![locale.languageCode] ?? '';
  String get finalizing => _localizedValues['finalizing']![locale.languageCode] ?? '';
  String get waitingConnection => _localizedValues['waitingConnection']![locale.languageCode] ?? '';
  String get saveTo => _localizedValues['saveTo']![locale.languageCode] ?? '';
  String get from => _localizedValues['from']![locale.languageCode] ?? '';

  String get manualConnectTitle => _localizedValues['manualConnectTitle']![locale.languageCode] ?? '';
  String get ipAddress => _localizedValues['ipAddress']![locale.languageCode] ?? '';
  String get ipHint => _localizedValues['ipHint']![locale.languageCode] ?? '';
  String get port => _localizedValues['port']![locale.languageCode] ?? '';
  String get next => _localizedValues['next']![locale.languageCode] ?? '';

  String get historyTitle => _localizedValues['historyTitle']![locale.languageCode] ?? '';
  String get noHistory => _localizedValues['noHistory']![locale.languageCode] ?? '';
  String get clearAll => _localizedValues['clearAll']![locale.languageCode] ?? '';
  String get deleteRecord => _localizedValues['deleteRecord']![locale.languageCode] ?? '';
  String get deleteFile => _localizedValues['deleteFile']![locale.languageCode] ?? '';
  String get deleteConfirmTitle => _localizedValues['deleteConfirmTitle']![locale.languageCode] ?? '';
  String get deleteFileConfirm => _localizedValues['deleteFileConfirm']![locale.languageCode] ?? '';
  String get clearAllConfirm => _localizedValues['clearAllConfirm']![locale.languageCode] ?? '';

  String get detailSend => _localizedValues['detailSend']![locale.languageCode] ?? '';
  String get detailReceive => _localizedValues['detailReceive']![locale.languageCode] ?? '';
  String get time => _localizedValues['time']![locale.languageCode] ?? '';
  String get peer => _localizedValues['peer']![locale.languageCode] ?? '';
  String get size => _localizedValues['size']![locale.languageCode] ?? '';
  String get duration => _localizedValues['duration']![locale.languageCode] ?? '';
  String get speed => _localizedValues['speed']![locale.languageCode] ?? '';
  String get fileCount => _localizedValues['fileCount']![locale.languageCode] ?? '';
  String get files => _localizedValues['files']![locale.languageCode] ?? '';
  String get savePath => _localizedValues['savePath']![locale.languageCode] ?? '';
  String get status => _localizedValues['status']![locale.languageCode] ?? '';
  String get success => _localizedValues['success']![locale.languageCode] ?? '';
  String get failed => _localizedValues['failed']![locale.languageCode] ?? '';
  String get expand => _localizedValues['expand']![locale.languageCode] ?? '';
  String get collapse => _localizedValues['collapse']![locale.languageCode] ?? '';
  String fileCountMore(int count) {
    final template = _localizedValues['fileCountMore']![locale.languageCode] ?? '... {count} files';
    return template.replaceAll('{count}', count.toString());
  }

  // 设置页面 getter（已在上面有部分，补充完整）
  String get settings => _localizedValues['settings']![locale.languageCode] ?? '';
  String get language => _localizedValues['language']![locale.languageCode] ?? '';
  String get chinese => _localizedValues['chinese']![locale.languageCode] ?? '';
  String get english => _localizedValues['english']![locale.languageCode] ?? '';
  String get storageSettings => _localizedValues['storageSettings']![locale.languageCode] ?? '';
  String get publicDownloads => _localizedValues['publicDownloads']![locale.languageCode] ?? '';
  String get publicDesc => _localizedValues['publicDesc']![locale.languageCode] ?? '';
  String get appPrivate => _localizedValues['appPrivate']![locale.languageCode] ?? '';
  String get privateDesc => _localizedValues['privateDesc']![locale.languageCode] ?? '';
  String get requestPermission => _localizedValues['requestPermission']![locale.languageCode] ?? '';
  String get transferSettings => _localizedValues['transferSettings']![locale.languageCode] ?? '';
  String get concurrency => _localizedValues['concurrency']![locale.languageCode] ?? '';
  String get maxChunk => _localizedValues['maxChunk']![locale.languageCode] ?? '';
  String get securitySettings => _localizedValues['securitySettings']![locale.languageCode] ?? '';
  String get tlsEncryption => _localizedValues['tlsEncryption']![locale.languageCode] ?? '';
  String get tlsDesc => _localizedValues['tlsDesc']![locale.languageCode] ?? '';
  String get permissionGranted => _localizedValues['permissionGranted']![locale.languageCode] ?? '';
  String get permissionDenied => _localizedValues['permissionDenied']![locale.languageCode] ?? '';

  String get selectSaveDir => _localizedValues['selectSaveDir']![locale.languageCode] ?? '';
  String get selectSaveDirDesc => _localizedValues['selectSaveDirDesc']![locale.languageCode] ?? '';
  String get cannotUsePublic => _localizedValues['cannotUsePublic']![locale.languageCode] ?? '';
  String get fallbackToPrivate => _localizedValues['fallbackToPrivate']![locale.languageCode] ?? '';
  String get noWritePermission => _localizedValues['noWritePermission']![locale.languageCode] ?? '';
  String get chooseAction => _localizedValues['chooseAction']![locale.languageCode] ?? '';
  String get useInternalStorage => _localizedValues['useInternalStorage']![locale.languageCode] ?? '';
  String get requestPermissionAndRetry => _localizedValues['requestPermissionAndRetry']![locale.languageCode] ?? '';
  String get switchedToInternal => _localizedValues['switchedToInternal']![locale.languageCode] ?? '';
  String get error => _localizedValues['error']![locale.languageCode] ?? '';
  String get remaining => _localizedValues['remaining']![locale.languageCode] ?? '';
}

class AppLocalizationsDelegate extends LocalizationsDelegate<AppLocalizations> {
  const AppLocalizationsDelegate();

  @override
  bool isSupported(Locale locale) => ['en', 'zh'].contains(locale.languageCode);

  @override
  Future<AppLocalizations> load(Locale locale) async {
    return AppLocalizations(locale);
  }

  @override
  bool shouldReload(covariant LocalizationsDelegate<AppLocalizations> old) => false;
}
