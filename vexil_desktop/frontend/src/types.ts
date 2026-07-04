export interface TransferView {
    task_id: string;
    state: string;
    percent: number;
    speed: string;
    sent: string;
    total: string;
    eta: string;
    error?: string;
}

export interface DeviceInfo {
    name: string;
    ip: string;
    port: number;
}

export interface HistoryEntry {
    time: string;
    direction: string;
    peer: string;
    peer_name?: string;
    files: number;
    file_names: string[];
    size: number;
    duration_sec: number;
    speed_mbps: number;
    success: boolean;
    save_path?: string;
}

export type Tab = "send" | "recv";