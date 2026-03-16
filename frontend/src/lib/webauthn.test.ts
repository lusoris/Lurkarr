import { describe, it, expect } from 'vitest';
import { bufferToBase64url, base64urlToBuffer } from '$lib/webauthn';

describe('bufferToBase64url', () => {
	it('encodes an empty buffer', () => {
		const buf = new ArrayBuffer(0);
		expect(bufferToBase64url(buf)).toBe('');
	});

	it('encodes a known byte sequence', () => {
		// "Hello" in bytes → base64url
		const bytes = new Uint8Array([72, 101, 108, 108, 111]);
		const result = bufferToBase64url(bytes.buffer);
		expect(result).toBe('SGVsbG8');
	});

	it('does not contain +, /, or = characters', () => {
		// Use bytes that would normally produce + / = in standard base64
		const bytes = new Uint8Array([255, 254, 253, 252, 251, 250]);
		const result = bufferToBase64url(bytes.buffer);
		expect(result).not.toContain('+');
		expect(result).not.toContain('/');
		expect(result).not.toContain('=');
	});

	it('roundtrips with base64urlToBuffer', () => {
		const original = new Uint8Array([0, 1, 127, 128, 255]);
		const encoded = bufferToBase64url(original.buffer);
		const decoded = new Uint8Array(base64urlToBuffer(encoded));
		expect(decoded).toEqual(original);
	});
});

describe('base64urlToBuffer', () => {
	it('decodes a known base64url string', () => {
		// "SGVsbG8" is base64url for "Hello"
		const buf = base64urlToBuffer('SGVsbG8');
		const bytes = new Uint8Array(buf);
		expect(Array.from(bytes)).toEqual([72, 101, 108, 108, 111]);
	});

	it('handles padding correctly', () => {
		// base64url with 1 char needing 3 pads -> should still work
		const buf = base64urlToBuffer('QQ');
		expect(new Uint8Array(buf)[0]).toBe(65); // 'A'
	});

	it('converts - back to + and _ back to /', () => {
		// Standard base64 "//8=" -> base64url "__8"
		const buf = base64urlToBuffer('__8');
		const bytes = new Uint8Array(buf);
		expect(bytes[0]).toBe(255);
		expect(bytes[1]).toBe(255);
	});

	it('roundtrips with bufferToBase64url', () => {
		const input = 'dGVzdGluZw';
		const buf = base64urlToBuffer(input);
		const output = bufferToBase64url(buf);
		expect(output).toBe(input);
	});

	it('decodes empty string', () => {
		const buf = base64urlToBuffer('');
		expect(new Uint8Array(buf)).toHaveLength(0);
	});
});
