import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import { GetLanguage, SetLanguage } from "../../wailsjs/go/gui/Handler";

type Lang = "zh" | "en";

interface I18nContextType {
  lang: Lang;
  t: (key: string, params?: Record<string, string | number>) => string;
  switchLang: (lang: Lang) => Promise<void>;
}

const I18nContext = createContext<I18nContextType>({
  lang: "zh",
  t: (key) => key,
  switchLang: async () => {},
});

const messages: Record<string, Record<Lang, string>> = {
  "app.title": { zh: "Vexil", en: "Vexil" },
  "button.got_it": { zh: "知道了", en: "Got it" },
  "button.cancel": { zh: "取消", en: "Cancel" },
  "button.save": { zh: "保存设置", en: "Save" },
  "button.saved": { zh: "✅ 已保存", en: "✅ Saved" },
  "button.delete_all": { zh: "清空全部记录", en: "Clear All Records" },
  "button.view_all": { zh: "查看全部", en: "View All" },
  "button.refresh": { zh: "刷新", en: "Refresh" },
  "button.clear": { zh: "清空", en: "Clear" },
  "button.manual_add": { zh: "+ 手动添加设备", en: "+ Add Device Manually" },
  "button.add": { zh: "添加", en: "Add" },
  "button.cancel_small": { zh: "取消", en: "Cancel" },
  "button.select_file": { zh: "📁 选择文件", en: "📁 Select File" },
  "button.select_folder": { zh: "📂 选择文件夹", en: "📂 Select Folder" },
  "button.start_send": { zh: "开始发送", en: "Start Sending" },
  "button.start_recv": { zh: "开始接收", en: "Start Receiving" },
  "button.open_file": { zh: "📂 打开文件", en: "📂 Open File" },
  "button.delete_file": { zh: "🗑 删除文件", en: "🗑 Delete File" },

  "theme.switch_light": { zh: "切换亮色主题", en: "Switch to Light" },
  "theme.switch_dark": { zh: "切换暗色主题", en: "Switch to Dark" },

  "guide.welcome": { zh: "欢迎使用 Vexil", en: "Welcome to Vexil" },
  "guide.send_desc": { zh: "📤 发送：选择文件 → 选择设备 → 开始发送", en: "📤 Send: Select files → Choose device → Start" },
  "guide.recv_desc": { zh: "📥 接收：选择保存位置 → 等待发送", en: "📥 Receive: Choose save folder → Wait for sender" },
  "guide.resume": { zh: "🔄 支持断点续传，关闭后自动恢复", en: "🔄 Supports resume, auto-recover after close" },
  "guide.checksum": { zh: "🔐 流式 SHA-256 校验，确保数据完整", en: "🔐 Stream SHA-256 checksum for data integrity" },
  "guide.theme": { zh: "🌓 右上角切换亮色/暗色主题", en: "🌓 Toggle light/dark theme in top right" },
  "guide.settings": { zh: "右上角齿轮进入设置", en: "Gear icon for settings" },

  "device.searching": { zh: "搜索中...", en: "Searching..." },
  "device.found_title": { zh: "发现的设备：", en: "Devices found:" },
  "device.no_device": { zh: "暂无设备", en: "No devices" },
  "device.select_device": { zh: "选择设备", en: "Select Device" },
  "device.ip_placeholder": { zh: "IP 地址", en: "IP Address" },
  "device.port_placeholder": { zh: "端口", en: "Port" },
  "device.error_port": { zh: "请输入有效端口号", en: "Enter a valid port" },
  "device.error_select_target": { zh: "请选择目标设备", en: "Please select a target device" },
  "device.error_select_files": { zh: "请选择文件或文件夹", en: "Please select files or folders" },

  "send.title": { zh: "发送", en: "Send" },
  "send.drag_placeholder": { zh: "拖拽文件到此处", en: "Drop files here" },
  "send.files_selected": { zh: "已选 {count} 个文件/文件夹，继续拖拽添加", en: "{count} file(s)/folder(s) selected, drop to add more" },
  "send.progress.waiting": { zh: "等待中", en: "Waiting" },
  "send.progress.transferring": { zh: "传输中...", en: "Transferring..." },
  "send.progress.finalizing": { zh: "校验中...", en: "Finalizing..." },
  "send.progress.completed": { zh: "✅ 完成", en: "✅ Completed" },
  "send.progress.failed": { zh: "❌ 失败", en: "❌ Failed" },
  "send.progress.cancelled": { zh: "已取消", en: "Cancelled" },

  "recv.title": { zh: "接收", en: "Receive" },
  "recv.save_dir": { zh: "保存位置", en: "Save Location" },
  "recv.port_label": { zh: "监听端口", en: "Listen Port" },
  "recv.progress.waiting": { zh: "等待连接...", en: "Waiting for connection..." },
  "recv.progress.receiving": { zh: "接收中...", en: "Receiving..." },
  "recv.progress.finalizing": { zh: "校验中...", en: "Finalizing..." },
  "recv.progress.completed": { zh: "✅ 完成", en: "✅ Completed" },
  "recv.progress.failed": { zh: "❌ 失败", en: "❌ Failed" },
  "recv.progress.cancelled": { zh: "已取消", en: "Cancelled" },

  "history.title": { zh: "传输记录", en: "Transfer History" },
  "history.no_records": { zh: "还没有传输记录", en: "No transfer records" },
  "history.recent": { zh: "最近 {count} 条传输记录", en: "Recent {count} transfer records" },
  "history.detail_title": { zh: "传输详情", en: "Transfer Details" },
  "history.detail.success": { zh: "成功", en: "Success" },
  "history.detail.failed": { zh: "失败", en: "Failed" },
  "history.detail.send_label": { zh: "↑ 发送", en: "↑ Sent" },
  "history.detail.recv_label": { zh: "↓ 接收", en: "↓ Received" },
  "history.detail.files": { zh: "文件", en: "Files" },
  "history.detail.size": { zh: "大小", en: "Size" },
  "history.detail.speed": { zh: "速度", en: "Speed" },
  "history.detail.duration": { zh: "耗时", en: "Duration" },
  "history.detail.peer_name": { zh: "对方", en: "Peer Name" },
  "history.detail.peer_ip": { zh: "设备", en: "Device" },
  "history.detail.time": { zh: "时间", en: "Time" },
  "history.detail.save_path": { zh: "路径", en: "Path" },

  "progress.preparing": { zh: "准备中...", en: "Preparing..." },
  "progress.connecting": { zh: "连接中...", en: "Connecting..." },
  "progress.running": { zh: "", en: "" },
  "progress.finalizing": { zh: "校验中...", en: "Finalizing..." },
  "progress.completed": { zh: "✅ 传输完成", en: "✅ Transfer Complete" },
  "progress.failed": { zh: "❌ 传输失败", en: "❌ Transfer Failed" },
  "progress.cancelled": { zh: "已取消", en: "Cancelled" },

  "settings.title": { zh: "⚙ 设置", en: "⚙ Settings" },
  "settings.port": { zh: "默认接收端口", en: "Default Receive Port" },
  "settings.save_dir": { zh: "默认保存目录", en: "Default Save Directory" },
  "settings.advanced": { zh: "高级选项", en: "Advanced Options" },
  "settings.conns": { zh: "并发连接数", en: "Concurrent Connections" },
  "settings.chunk_mb": { zh: "块大小 (MB)", en: "Chunk Size (MB)" },
  "settings.window_mb": { zh: "窗口大小 (MB)", en: "Window Size (MB)" },
  "settings.tls": { zh: "TLS 加密", en: "TLS Encryption" },
  "settings.tls_enabled": { zh: "已启用", en: "Enabled" },
  "settings.tls_disabled": { zh: "已禁用", en: "Disabled" },
  "settings.language": { zh: "语言", en: "Language" },
  "settings.lang_zh": { zh: "中文", en: "Chinese" },
  "settings.lang_en": { zh: "English", en: "English" },
};

export const LanguageProvider = ({ children }: { children: ReactNode }) => {
  const [lang, setLang] = useState<Lang>("zh");

  useEffect(() => {
    (async () => {
      try {
        const l = await GetLanguage();
        if (l === "en") setLang("en");
      } catch {}
    })();
  }, []);

  const switchLang = useCallback(async (newLang: Lang) => {
    try {
      await SetLanguage(newLang);
      setLang(newLang);
    } catch (e) {
      console.error("Failed to set language", e);
    }
  }, []);

  const t = useCallback((key: string, params?: Record<string, string | number>) => {
    const entry = messages[key];
    if (!entry) return key;
    const text = entry[lang] || entry["zh"] || key;
    if (params) {
      return text.replace(/\{(\w+)\}/g, (_, k) => String(params[k] ?? `{${k}}`));
    }
    return text;
  }, [lang]);

  return (
    <I18nContext.Provider value={{ lang, t, switchLang }}>
      {children}
    </I18nContext.Provider>
  );
};

export const useLanguage = () => useContext(I18nContext);