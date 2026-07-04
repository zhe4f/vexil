import type { TransferView } from "../types";
import { useLanguage } from "../i18n/i18n";

interface Props {
  view: TransferView;
}

export function ProgressBar({ view }: Props) {
  const { t } = useLanguage();
  const stateLabel = () => {
    switch (view.state) {
      case "preparing": return t("progress.preparing");
      case "connecting": return t("progress.connecting");
      case "running": return t("progress.running");
      case "finalizing": return t("progress.finalizing");
      case "completed": return t("progress.completed");
      case "failed": return t("progress.failed");
      case "cancelled": return t("progress.cancelled");
      default: return view.state;
    }
  };

  const label = stateLabel();
  const isRunning = view.state === "running" || view.state === "finalizing" || view.state === "";
  const isDone = view.state === "completed" || view.state === "failed" || view.state === "cancelled";

  return (
    <div style={{ width: "100%", textAlign: "center", marginTop: 24 }}>
      {label && <p style={{ fontSize: 15, color: "#8b949e", marginBottom: 12 }}>{label}</p>}
      {view.error && <p style={{ fontSize: 13, color: "#f85149", marginBottom: 12 }}>{view.error}</p>}
      {(isRunning || isDone) && (
        <>
          <div style={{ width: "100%", height: 6, background: "rgba(255,255,255,0.06)", borderRadius: 3, overflow: "hidden" }}>
            <div style={{
              width: `${view.percent}%`, height: "100%",
              background: isDone ? "#3fb950" : "linear-gradient(90deg, #3b82f6, #6366f1)",
              borderRadius: 3, transition: "width 0.3s ease",
            }} />
          </div>
          <div style={{ display: "flex", justifyContent: "space-between", marginTop: 10, fontSize: 13, color: "#8b949e" }}>
            <span>{view.percent.toFixed(1)}%</span>
            <span>{view.sent} / {view.total}</span>
            <span>{view.speed}</span>
            <span>{view.eta}</span>
          </div>
        </>
      )}
    </div>
  );
}