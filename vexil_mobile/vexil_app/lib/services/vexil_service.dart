import 'dart:async';
import 'dart:convert';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/device_info.dart';
import '../models/history_entry.dart';

class VexilService {
  static const _methodChannel = MethodChannel('com.vexil/vexil');
  static const _eventChannel = EventChannel('com.vexil/vexil_events');

  Stream<Map<String, dynamic>>? _eventStream;

  Stream<Map<String, dynamic>> get events {
    _eventStream ??= _eventChannel.receiveBroadcastStream().map((event) {
      final map = Map<String, dynamic>.from(event as Map);
      return map;
    });
    return _eventStream!;
  }

  Future<String> _getDeviceName() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString('device_name') ?? '';
  }

  Future<List<DeviceInfo>> discoverDevices({int timeoutSec = 3}) async {
    await Future.delayed(const Duration(milliseconds: 50));
    try {
      final result = await _methodChannel.invokeMethod('discoverDevices', {
        'timeout': timeoutSec,
      });
      if (result == null) return [];
      String jsonStr;
      if (result is String) {
        jsonStr = result;
      } else {
        jsonStr = result.toString();
      }
      if (jsonStr.isEmpty || jsonStr == 'null' || jsonStr == '[]') {
        return [];
      }
      final list = jsonDecode(jsonStr) as List<dynamic>;
      return list.map((e) => DeviceInfo.fromJson(e as Map<String, dynamic>)).toList();
    } catch (e) {
      return [];
    }
  }

  Future<String> sendFiles(String host, int port, List<String> paths, String peerName) async {
    final pathsJson = jsonEncode(paths);
    final deviceName = await _getDeviceName();
    final prefs = await SharedPreferences.getInstance();
    return await _methodChannel.invokeMethod('sendFiles', {
      'host': host,
      'port': port,
      'paths': pathsJson,
      'peerName': deviceName.isNotEmpty ? deviceName : peerName,
      'numConns': prefs.getInt('num_conns') ?? 4,
      'maxChunkMB': prefs.getInt('max_chunk_mb') ?? 16,
      'tlsEnabled': prefs.getBool('tls_enabled') ?? true,
    });
  }

  Future<String> startReceive(int port, String saveDir) async {
    final deviceName = await _getDeviceName();
    final prefs = await SharedPreferences.getInstance();
    return await _methodChannel.invokeMethod('startReceive', {
      'port': port,
      'saveDir': saveDir,
      'deviceName': deviceName,
      'numConns': prefs.getInt('num_conns') ?? 4,
      'maxChunkMB': prefs.getInt('max_chunk_mb') ?? 16,
      'tlsEnabled': prefs.getBool('tls_enabled') ?? true,
    });
  }

  Future<void> cancelTransfer(String taskId) async {
    await _methodChannel.invokeMethod('cancelTransfer', {'taskId': taskId});
  }

  Future<List<HistoryEntry>> getHistory({int limit = 20}) async {
    try {
      final result = await _methodChannel.invokeMethod('getHistory', {'limit': limit});
      print('[History] getHistory result: $result');
      if (result == null || result.toString() == 'null') return [];
      final list = jsonDecode(result.toString()) as List<dynamic>;
      return list.map((e) => HistoryEntry.fromJson(e as Map<String, dynamic>)).toList();
    } catch (e) {
      print('[History] getHistory error: $e');
      return [];
    }
  }

  Future<void> clearHistory() async {
    await _methodChannel.invokeMethod('clearHistory');
  }

  Future<void> deleteHistory(int index) async {
    await _methodChannel.invokeMethod('deleteHistory', {'index': index});
  }

  Future<void> openFile(String path) async {
    await _methodChannel.invokeMethod('openFile', {'path': path});
  }

  //delete receive file
  Future<void> deleteFile(String savePath, List<String> fileNames) async {
    await _methodChannel.invokeMethod('deleteFile', {
      'savePath': savePath,
      'fileNames': fileNames,
    });
  }

  //delete send file
  Future<void> deleteFiles(String pathsStr) async {
    await _methodChannel.invokeMethod('deleteFiles', {'paths': pathsStr});
  }
  
  //悬浮窗
  Future<void> startForegroundService() async {
    await _methodChannel.invokeMethod('startForegroundService');
  }

  Future<void> stopForegroundService() async {
    await _methodChannel.invokeMethod('stopForegroundService');
  }

  Future<void> updateNotification(String title, String content, {bool stopAfter = false}) async {
    await _methodChannel.invokeMethod('updateNotification', {
      'title': title,
      'content': content,
      'stopAfter': stopAfter,
    });
  }
}