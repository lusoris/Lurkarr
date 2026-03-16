import { describe, it, expect, vi, beforeEach } from 'vitest';

// vi.hoisted runs before vi.mock's hoisted factory, so mockToast is available.
const mockToast = vi.hoisted(() => ({
	success: vi.fn(),
	error: vi.fn(),
	info: vi.fn(),
	warning: vi.fn()
}));
vi.mock('svelte-sonner', () => ({ toast: mockToast }));

import { getToasts } from '$lib/stores/toast.svelte';

describe('toast store', () => {
	let toasts: ReturnType<typeof getToasts>;

	beforeEach(() => {
		vi.clearAllMocks();
		toasts = getToasts();
	});

	it('success calls toast.success', () => {
		toasts.success('done');
		expect(mockToast.success).toHaveBeenCalledWith('done');
	});

	it('error calls toast.error', () => {
		toasts.error('fail');
		expect(mockToast.error).toHaveBeenCalledWith('fail');
	});

	it('info calls toast.info', () => {
		toasts.info('heads up');
		expect(mockToast.info).toHaveBeenCalledWith('heads up');
	});

	it('warning calls toast.warning', () => {
		toasts.warning('careful');
		expect(mockToast.warning).toHaveBeenCalledWith('careful');
	});

	it('returns an object with all four methods', () => {
		expect(typeof toasts.success).toBe('function');
		expect(typeof toasts.error).toBe('function');
		expect(typeof toasts.info).toBe('function');
		expect(typeof toasts.warning).toBe('function');
	});
});
