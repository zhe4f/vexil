package i18n

// T 根据 lang 返回 key 对应的翻译字符串，lang 为空时默认中文
func T(lang, key string) string {
	if lang == "" {
		lang = "zh"
	}
	if m, ok := messages[key]; ok {
		if str, ok := m[lang]; ok {
			return str
		}
	}
	// 回退中文
	if m, ok := messages[key]; ok {
		if str, ok := m["zh"]; ok {
			return str
		}
	}
	return key
}

var messages = map[string]map[string]string{
	"usage": {
		"zh": `vexil send <host:port> <file/dir...>
vexil recv <port> [saveDir]
vexil name [newName]          查看或设置设备名
vexil lang [zh/en]            查看或设置语言
vexil discover [timeout_seconds]
vexil history [count]
vexil history clear           清空全部记录
vexil history clear <序号>     删除指定记录`,
		"en": `vexil send <host:port> <file/dir...>
vexil recv <port> [saveDir]
vexil name [newName]          View or set device name
vexil lang [zh/en]            View or set language
vexil discover [timeout_seconds]
vexil history [count]
vexil history clear           Clear all records
vexil history clear <index>   Delete specified record`,
	},
	"send_usage": {
		"zh": "vexil send <host:port> <file/dir...>",
		"en": "vexil send <host:port> <file/dir...>",
	},
	"invalid_port": {
		"zh": "无效的端口号: %s",
		"en": "Invalid port: %s",
	},
	"no_files": {
		"zh": "请指定至少一个文件或目录",
		"en": "Please specify at least one file or directory",
	},
	"searching_device": {
		"zh": "  正在查找设备 %s...\n",
		"en": "  Searching for device %s...\n",
	},
	"device_found": {
		"zh": "  找到: %s (%s:%d)\n",
		"en": "  Found: %s (%s:%d)\n",
	},
	"unresolvable_host": {
		"zh": "无法解析主机名，请使用 IP:端口 或先运行 vexil discover 查找设备",
		"en": "Unable to resolve hostname, use IP:port or run vexil discover first",
	},
	"task_already_connected": {
		"zh": "已有传输连接到 %s",
		"en": "Already connected to %s",
	},
	"task_not_found": {
		"zh": "任务不存在: %s",
		"en": "Task not found: %s",
	},
	"task_error": {
		"zh": "error: %v",
		"en": "error: %v",
	},
	"task_create_failed": {
		"zh": "error: 任务创建失败",
		"en": "error: Failed to create task",
	},
	"done": {
		"zh": "\ndone",
		"en": "\ndone",
	},
	"cancelled": {
		"zh": "\ncancelled",
		"en": "\ncancelled",
	},
	"failed": {
		"zh": "\nfailed",
		"en": "\nfailed",
	},
	"signal_received": {
		"zh": "\n  收到信号 %v，正在取消...\n",
		"en": "\n  Signal %v received, cancelling...\n",
	},
	"recv_usage": {
		"zh": "vexil recv <port> [saveDir]",
		"en": "vexil recv <port> [saveDir]",
	},
	"udp_start_fail": {
		"zh": "警告: UDP 发现启动失败: %v\n",
		"en": "Warning: UDP discovery start failed: %v\n",
	},
	"mdns_start_fail": {
		"zh": "警告: mDNS 发现启动失败: %v\n",
		"en": "Warning: mDNS discovery start failed: %v\n",
	},
	"listening": {
		"zh": "  监听端口: %d, 保存目录: %s\n",
		"en": "  Listening on port: %d, save directory: %s\n",
	},
	"discovering": {
		"zh": "  正在发现设备 (%v)...\n\n",
		"en": "  Discovering devices (%v)...\n\n",
	},
	"discover_fail": {
		"zh": "  发现失败: %v\n",
		"en": "  Discovery failed: %v\n",
	},
	"no_devices": {
		"zh": "  未发现设备\n",
		"en": "  No devices found\n",
	},
	"no_history": {
		"zh": "暂无传输记录\n",
		"en": "No transfer records\n",
	},
	"recent_history": {
		"zh": "  最近 %d 条传输记录:\n\n",
		"en": "  Recent %d transfer records:\n\n",
	},
	"history_cleared": {
		"zh": "已清空全部传输记录\n",
		"en": "All transfer records cleared\n",
	},
	"history_deleted": {
		"zh": "已删除第 %d 条记录\n",
		"en": "Deleted record #%d\n",
	},
	"clear_fail": {
		"zh": "清空失败: %v\n",
		"en": "Clear failed: %v\n",
	},
	"delete_fail": {
		"zh": "删除失败: %v\n",
		"en": "Delete failed: %v\n",
	},
	"invalid_index": {
		"zh": "无效的序号: %s\n",
		"en": "Invalid index: %s\n",
	},
	"current_device_name": {
		"zh": "当前设备名: %s\n",
		"en": "Current device name: %s\n",
	},
	"device_name_default": {
		"zh": "当前设备名: %s (系统默认)\n",
		"en": "Current device name: %s (system default)\n",
	},
	"device_name_updated": {
		"zh": "设备名已更新为: %s\n",
		"en": "Device name updated to: %s\n",
	},
	"name_save_fail": {
		"zh": "保存失败: %v\n",
		"en": "Save failed: %v\n",
	},
	"current_lang": {
		"zh": "当前语言: %s\n",
		"en": "Current language: %s\n",
	},
	"lang_updated": {
		"zh": "语言已更新为: %s\n",
		"en": "Language updated to: %s\n",
	},
	"lang_invalid": {
		"zh": "无效的语言代码，仅支持 zh 或 en\n",
		"en": "Invalid language code, only zh or en supported\n",
	},
	"transfer_complete": {
		"zh": "传输完成",
		"en": "Transfer completed",
	},
	"transfer_failed": {
		"zh": "传输失败",
		"en": "Transfer failed",
	},
	"transfer_cancelled": {
		"zh": "传输已取消",
		"en": "Transfer cancelled",
	},
	"direction_send": {
		"zh": "↑ 发送至",
		"en": "↑ Sent to",
	},
	"direction_recv": {
		"zh": "↓ 接收自",
		"en": "↓ Received from",
	},
}