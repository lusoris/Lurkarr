import { describe, it, expect } from 'vitest';
import { cn } from '$lib/lib/utils';

describe('cn', () => {
	it('merges class strings', () => {
		expect(cn('px-2', 'py-2')).toBe('px-2 py-2');
	});

	it('handles conditional classes via clsx', () => {
		expect(cn('base', false && 'hidden', 'extra')).toBe('base extra');
	});

	it('merges conflicting tailwind classes (last wins)', () => {
		// tailwind-merge should resolve conflicts
		expect(cn('px-2', 'px-4')).toBe('px-4');
	});

	it('handles undefined and null values', () => {
		expect(cn('a', undefined, null, 'b')).toBe('a b');
	});

	it('handles empty inputs', () => {
		expect(cn()).toBe('');
		expect(cn('')).toBe('');
	});

	it('handles array inputs', () => {
		expect(cn(['px-2', 'py-2'])).toBe('px-2 py-2');
	});

	it('handles object inputs', () => {
		expect(cn({ 'px-2': true, hidden: false })).toBe('px-2');
	});

	it('deduplicates classes', () => {
		expect(cn('px-2 py-2', 'px-2')).toBe('py-2 px-2');
	});

	it('resolves responsive prefix conflicts', () => {
		expect(cn('sm:px-2', 'sm:px-4')).toBe('sm:px-4');
	});
});
