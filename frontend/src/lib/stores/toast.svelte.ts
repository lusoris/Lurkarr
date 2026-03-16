import { toast } from 'svelte-sonner';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export function getToasts() {
	return {
		success: (msg: string) => toast.success(msg),
		error: (msg: string) => toast.error(msg),
		info: (msg: string) => toast.info(msg),
		warning: (msg: string) => toast.warning(msg)
	};
}
