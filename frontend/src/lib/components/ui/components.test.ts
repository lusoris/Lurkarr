import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import { createRawSnippet } from 'svelte';
import Badge from '$lib/components/ui/Badge.svelte';
import Button from '$lib/components/ui/Button.svelte';
import Card from '$lib/components/ui/Card.svelte';
import Input from '$lib/components/ui/Input.svelte';
import Toggle from '$lib/components/ui/Toggle.svelte';

afterEach(() => cleanup());

function textSnippet(text: string) {
	return createRawSnippet(() => ({
		render: () => `<span>${text}</span>`
	}));
}

describe('Badge', () => {
	it('renders children text', () => {
		render(Badge, { props: { children: textSnippet('Active') } });
		expect(screen.getByText('Active')).toBeInTheDocument();
	});

	it('applies default variant classes', () => {
		render(Badge, { props: { children: textSnippet('Tag') } });
		const el = screen.getByText('Tag').closest('span.inline-flex');
		expect(el!.className).toContain('bg-surface-700');
	});

	it('applies success variant classes', () => {
		render(Badge, { props: { variant: 'success', children: textSnippet('OK') } });
		const el = screen.getByText('OK').closest('span.inline-flex');
		expect(el!.className).toContain('bg-lurk-900/50');
	});

	it('applies error variant classes', () => {
		render(Badge, { props: { variant: 'error', children: textSnippet('Fail') } });
		const el = screen.getByText('Fail').closest('span.inline-flex');
		expect(el!.className).toContain('bg-red-900/50');
	});
});

describe('Button', () => {
	it('renders children text', () => {
		render(Button, { props: { children: textSnippet('Click me') } });
		expect(screen.getByText('Click me')).toBeInTheDocument();
	});

	it('calls onclick when clicked', async () => {
		const onclick = vi.fn();
		render(Button, { props: { onclick, children: textSnippet('Go') } });
		await fireEvent.click(screen.getByRole('button'));
		expect(onclick).toHaveBeenCalledOnce();
	});

	it('disables the button when disabled=true', () => {
		render(Button, { props: { disabled: true, children: textSnippet('Nope') } });
		expect(screen.getByRole('button')).toBeDisabled();
	});

	it('applies danger variant classes', () => {
		render(Button, { props: { variant: 'danger', children: textSnippet('Delete') } });
		const btn = screen.getByRole('button');
		expect(btn.className).toContain('bg-red-600');
	});

	it('applies size classes', () => {
		render(Button, { props: { size: 'lg', children: textSnippet('Big') } });
		const btn = screen.getByRole('button');
		expect(btn.className).toContain('px-6');
	});

	it('shows spinner when loading', () => {
		render(Button, { props: { loading: true, children: textSnippet('Wait') } });
		const btn = screen.getByRole('button');
		const svg = btn.querySelector('svg');
		expect(svg).not.toBeNull();
		expect(svg!.classList.contains('animate-spin')).toBe(true);
	});
});

describe('Card', () => {
	it('renders children', () => {
		render(Card, { props: { children: textSnippet('Content') } });
		expect(screen.getByText('Content')).toBeInTheDocument();
	});

	it('applies custom class', () => {
		render(Card, { props: { class: 'my-custom', children: textSnippet('Hi') } });
		const el = screen.getByText('Hi').closest('div');
		expect(el!.className).toContain('my-custom');
	});

	it('gets button role when onclick provided', async () => {
		const onclick = vi.fn();
		render(Card, { props: { onclick, children: textSnippet('Click card') } });
		const el = screen.getByRole('button');
		expect(el).toBeInTheDocument();
		await fireEvent.click(el);
		expect(onclick).toHaveBeenCalledOnce();
	});
});

describe('Input', () => {
	it('renders with label', () => {
		render(Input, { props: { label: 'Email' } });
		expect(screen.getByText('Email')).toBeInTheDocument();
	});

	it('renders with placeholder', () => {
		render(Input, { props: { placeholder: 'Enter text' } });
		expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
	});

	it('shows error text', () => {
		render(Input, { props: { error: 'Required' } });
		expect(screen.getByText('Required')).toBeInTheDocument();
	});

	it('shows hint when no error', () => {
		render(Input, { props: { hint: 'Optional field' } });
		expect(screen.getByText('Optional field')).toBeInTheDocument();
	});

	it('disables input', () => {
		render(Input, { props: { disabled: true, placeholder: 'disabled' } });
		expect(screen.getByPlaceholderText('disabled')).toBeDisabled();
	});
});

describe('Toggle', () => {
	it('renders as a switch', () => {
		render(Toggle);
		expect(screen.getByRole('switch')).toBeInTheDocument();
	});

	it('renders label text', () => {
		render(Toggle, { props: { label: 'Dark mode' } });
		expect(screen.getByText('Dark mode')).toBeInTheDocument();
	});

	it('toggles checked on click', async () => {
		const onchange = vi.fn();
		render(Toggle, { props: { checked: false, onchange } });
		await fireEvent.click(screen.getByRole('switch'));
		expect(onchange).toHaveBeenCalledWith(true);
	});

	it('disables the switch', () => {
		render(Toggle, { props: { disabled: true } });
		expect(screen.getByRole('switch')).toBeDisabled();
	});

	it('renders hint text', () => {
		render(Toggle, { props: { hint: 'Enable feature' } });
		expect(screen.getByText('Enable feature')).toBeInTheDocument();
	});
});
