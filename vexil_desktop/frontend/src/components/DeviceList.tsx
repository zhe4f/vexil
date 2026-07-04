import { useLanguage } from "../i18n/i18n";
import type { DeviceInfo } from "../types";

interface Props {
  devices: DeviceInfo[];
  onSelect: (device: DeviceInfo) => void;
  loading: boolean;
}

export function DeviceList({ devices, onSelect, loading }: Props) {
  const { t } = useLanguage();
  if (loading) {
    return <p style={{ color: "#8b949e", fontSize: 13, textAlign: "center" }}>{t("device.searching")}</p>;
  }
  if (devices.length === 0) return null;

  return (
    <div style={{ marginTop: 12 }}>
      <p style={{ fontSize: 12, color: "#8b949e", marginBottom: 6 }}>{t("device.found_title")}</p>
      {devices.map((d, i) => (
        <button
          key={i}
          onClick={() => onSelect(d)}
          style={{
            display: "block", width: "100%", padding: "8px 12px", marginBottom: 4,
            background: "rgba(255,255,255,0.04)", border: "1px solid rgba(255,255,255,0.06)",
            borderRadius: 8, color: "#e6edf3", fontSize: 13, cursor: "pointer", textAlign: "left",
          }}
        >
          {d.name} ({d.ip}:{d.port})
        </button>
      ))}
    </div>
  );
}