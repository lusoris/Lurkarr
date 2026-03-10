interface LogMessage {
	id: number;
	app_type: string;
	level: string;
	message: string;
	created_at: string;
}

let messages = $state<LogMessage[]>([]);
let connected = $state(false);
let ws: WebSocket | null = null;
const MAX_MESSAGES = 1000;

export function getLogStream() {
	function connect() {
		const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
		ws = new WebSocket(`${proto}//${location.host}/ws/logs`);

		ws.onopen = () => { connected = true; };
		ws.onclose = () => {
			connected = false;
			setTimeout(connect, 3000);
		};
		ws.onmessage = (ev) => {
			try {
				const msg: LogMessage = JSON.parse(ev.data);
				messages = [...messages.slice(-(MAX_MESSAGES - 1)), msg];
			} catch { /* ignore parse errors */ }
		};
	}

	function disconnect() {
		ws?.close();
		ws = null;
	}

	function clear() {
		messages = [];
	}

	return {
		get messages() { return messages; },
		get connected() { return connected; },
		connect,
		disconnect,
		clear
	};
}
