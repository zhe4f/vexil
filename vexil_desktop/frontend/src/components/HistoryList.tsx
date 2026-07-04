import type { HistoryEntry } from "../types";
import { useLanguage } from "../i18n/i18n";

interface Props {
  entries: HistoryEntry[];
  onDelete?: (index: number) => void;
  onOpen?: (entry: HistoryEntry) => void;
  max?: number;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + " MB";
  return (bytes / 1024 / 1024 / 1024).toFixed(1) + " GB";
}

function formatTime(timeStr: string, lang: string): string {
  const date = new Date(timeStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  if (diff < 60000) return lang === "en" ? "Just now" : "刚刚";
  if (diff < 3600000) return `${Math.floor(diff / 60000)}${lang === "en" ? "m ago" : "分钟前"}`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}${lang === "en" ? "h ago" : "小时前"}`;
  return date.toLocaleDateString(lang === "en" ? "en-US" : "zh-CN");
}

export function HistoryList({ entries, onDelete, onOpen, max }: Props) {
  const { t, lang } = useLanguage();
  const list = max ? entries.slice(0, max) : entries;

  if (list.length === 0) {
    return <p style={{ color: "var(--text-muted)", fontSize: 13, textAlign: "center", padding: 12 }}>{t("history.no_records")}</p>;
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
      {list.map((entry, i) => (
        <div
          key={i}
          onClick={() => onOpen?.(entry)}
          style={{
            display: "flex", alignItems: "center", gap: 10,
            padding: "10px 14px", borderRadius: 10, cursor: onOpen ? "pointer" : "default",
            background: "var(--bg-input)", border: "1px solid var(--border)",
            transition: "background 0.2s",
          }}
          onMouseEnter={e => (e.currentTarget.style.background = "var(--bg-hover)")}
          onMouseLeave={e => (e.currentTarget.style.background = "var(--bg-input)")}
        >
          <span style={{ fontSize: 18, flexShrink: 0, color: entry.direction === "send" ? "var(--accent2)" : "var(--accent)" }}>
            {entry.direction === "send" ? "↑" : "↓"}
          </span>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 14, color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              {entry.file_names?.[0] || entry.files + "个文件"}
              {entry.files > 1 && <span style={{ color: "var(--text-secondary)", fontSize: 12 }}> +{entry.files - 1}</span>}
            </div>
            <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>
              {entry.peer_name || entry.peer} · {formatSize(entry.size)} · {formatTime(entry.time, lang)}
            </div>
          </div>
          <span style={{ fontSize: 16, flexShrink: 0, color: entry.success ? "var(--success)" : "var(--danger)" }}>
            {entry.success ? "✓" : "✗"}
          </span>
          {onDelete && (
            <button
              onClick={e => { e.stopPropagation(); onDelete(i + 1); }}
              style={{ background: "none", border: "none", color: "var(--text-muted)", fontSize: 14, cursor: "pointer", padding: 4 }}
            >✕</button>
          )}
        </div>
      ))}
    </div>
  );
}