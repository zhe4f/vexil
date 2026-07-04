enum TaskState { preparing, connecting, running, finalizing, completed, failed, cancelled }

class TransferTask {
  final String taskId;
  final TaskState state;
  final double percent;
  final double speedMBps;
  final int sent;
  final int total;
  final String? error;

  const TransferTask({
    required this.taskId,
    this.state = TaskState.preparing,
    this.percent = 0,
    this.speedMBps = 0,
    this.sent = 0,
    this.total = 0,
    this.error,
  });

  TransferTask copyWith({
    TaskState? state,
    double? percent,
    double? speedMBps,
    int? sent,
    int? total,
    String? error,
  }) {
    return TransferTask(
      taskId: taskId,
      state: state ?? this.state,
      percent: percent ?? this.percent,
      speedMBps: speedMBps ?? this.speedMBps,
      sent: sent ?? this.sent,
      total: total ?? this.total,
      error: error,
    );
  }
}