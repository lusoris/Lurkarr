import { describe, it, expect, vi, beforeEach } from 'vitest';

// The toast store uses $state (Svelte 5 runes) so this import goes through
// the Svelte compiler via the sveltekit vite plugin.
import { getToasts } from '$lib/stores/toast.svelte';

describe('toast store', () => {
	let toast: ReturnType<typeof getToasts>;

	beforeEach(() => {
		toast = getToasts();
		// Clear any existing toasts from previous tests
		for (const t of toast.items) {
			toast.remove(t.id);
		}
		vi.useFakeTimers();
	});

	it('starts empty', () => {
		expect(toast.items).toEqual([]);
	});

	it('success adds a toast', () => {
		toast.success('done');
		expect(toast.items).toHaveLength(1);
		expect(toast.items[0].type).toBe('success');
		expect(toast.items[0].message).toBe('done');
	});

	it('error adds a toast', () => {
		toast.error('fail');
		expect(toast.items).toHaveLength(1);
		expect(toast.items[0].type).toBe('error');
		expect(toast.items[0].message).toBe('fail');
	});

	it('info and warning add toasts', () => {
		toast.info('heads up');
		toast.warning('careful');
		expect(toast.items).toHaveLength(2);
		expect(toast.items[0].type).toBe('info');
		expect(toast.items[1].type).toBe('warning');
	});

	it('remove removes a specific toast', () => {
		toast.success('a');
		toast.success('b');
		const id = toast.items[0].id;
		toast.remove(id);
		expect(toast.items).toHaveLength(1);
		expect(toast.items[0].message).toBe('b');
	});

	it('auto-removes success after default duration', () => {
		toast.success('auto');
		expect(toast.items).toHaveLength(1);
		vi.advanceTimersByTime(4000);
		expect(toast.items).toHaveLength(0);
	});

	it('error auto-removes after 6 seconds', () => {
		toast.error('err');
		vi.advanceTimersByTime(4000);
		expect(toast.items).toHaveLength(1); // still there at 4s
		vi.advanceTimersByTime(2000);
		expect(toast.items).toHaveLength(0); // gone at 6s
	});

	it('each toast gets a unique id', () => {
		toast.success('a');
		toast.success('b');
		toast.success('c');
		const ids = toast.items.map((t) => t.id);
		expect(new Set(ids).size).toBe(3);
	});
});
