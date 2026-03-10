export type ToastType = 'success' | 'error' | 'info' | 'warning';

interface Toast {
	id: number;
	type: ToastType;
	message: string;
}

let toasts = $state<Toast[]>([]);
let nextId = 0;

export function getToasts() {
	function add(type: ToastType, message: string, duration = 4000) {
		const id = nextId++;
		toasts = [...toasts, { id, type, message }];
		if (duration > 0) {
			setTimeout(() => remove(id), duration);
		}
	}

	function remove(id: number) {
		toasts = toasts.filter((t) => t.id !== id);
	}

	return {
		get items() { return toasts; },
		success: (msg: string) => add('success', msg),
		error: (msg: string) => add('error', msg, 6000),
		info: (msg: string) => add('info', msg),
		warning: (msg: string) => add('warning', msg),
		remove
	};
}
