import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/device_info.dart';
import '../services/vexil_service.dart';

final vexilServiceProvider = Provider<VexilService>((ref) => VexilService());

final discoveryStateProvider = StateNotifierProvider<DiscoveryNotifier, AsyncValue<List<DeviceInfo>>>((ref) {
  return DiscoveryNotifier(ref.watch(vexilServiceProvider));
});

class DiscoveryNotifier extends StateNotifier<AsyncValue<List<DeviceInfo>>> {
  final VexilService _service;

  DiscoveryNotifier(this._service) : super(const AsyncValue.loading()) {
    scan();
  }

  Future<void> scan() async {
    state = const AsyncValue.loading();
    try {
      final devices = await _service.discoverDevices();
      state = AsyncValue.data(devices);
    } catch (e) {
      state = AsyncValue.error(e, StackTrace.current);
    }
  }
}