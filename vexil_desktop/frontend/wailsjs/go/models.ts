export namespace gui {
	
	export class ConfigInfo {
	    num_conns: number;
	    max_chunk_mb: number;
	    window_size_mb: number;
	    tls_enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConfigInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.num_conns = source["num_conns"];
	        this.max_chunk_mb = source["max_chunk_mb"];
	        this.window_size_mb = source["window_size_mb"];
	        this.tls_enabled = source["tls_enabled"];
	    }
	}
	export class DeviceInfo {
	    name: string;
	    ip: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	    }
	}
	export class HistoryEntry {
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
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.direction = source["direction"];
	        this.peer = source["peer"];
	        this.peer_name = source["peer_name"];
	        this.files = source["files"];
	        this.file_names = source["file_names"];
	        this.size = source["size"];
	        this.duration_sec = source["duration_sec"];
	        this.speed_mbps = source["speed_mbps"];
	        this.success = source["success"];
	        this.save_path = source["save_path"];
	    }
	}
	export class TLSFingerprint {
	    sha256: string;
	
	    static createFrom(source: any = {}) {
	        return new TLSFingerprint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sha256 = source["sha256"];
	    }
	}
	export class TaskSummary {
	    task_id: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task_id = source["task_id"];
	        this.state = source["state"];
	    }
	}

}

