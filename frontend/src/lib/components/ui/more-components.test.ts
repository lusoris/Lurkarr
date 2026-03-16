import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import { createRawSnippet } from 'svelte';
import Select from '$lib/components/ui/Select.svelte';
import Tabs from '$lib/components/ui/Tabs.svelte';
import Skeleton from '$lib/components/ui/Skeleton.svelte';
import EmptyState from '$lib/components/ui/EmptyState.svelte';
import Separator from '$lib/components/ui/Separator.svelte';
import PageHeader from '$lib/components/ui/PageHeader.svelte';
import DataTable from '$lib/components/ui/DataTable.svelte';

afterEach(() => cleanup());

function textSnippet(text: string) {
	return createRawSnippet(() => ({
		render: () => `<span>${text}</span>`
	}));
}

function optionsSnippet() {
	return createRawSnippet(() => ({
		render: () => `<option value="a">Alpha</option>`
	}));
}

// =============================================================================
// Select
// =============================================================================
describe('Select', () => {
	it('renders with label', () => {
		render(Select, { props: { label: 'Pick one', children: optionsSnippet() } });
		expect(screen.getByText('Pick one')).toBeInTheDocument();
	});

	it('renders a select (combobox) element', () => {
		render(Select, { props: { children: optionsSnippet() } });
		expect(screen.getByRole('combobox')).toBeInTheDocument();
	});

	it('renders hint text', () => {
		render(Select, { props: { hint: 'Choose wisely', children: optionsSnippet() } });
		expect(screen.getByText('Choose wisely')).toBeInTheDocument();
	});

	it('does not render label when not provided', () => {
		const { container } = render(Select, { props: { children: optionsSnippet() } });
		expect(container.querySelector('span.text-sm.font-medium')).toBeNull();
	});

	it('does not render hint when not provided', () => {
		const { container } = render(Select, { props: { children: optionsSnippet() } });
		expect(container.querySelector('p.text-xs')).toBeNull();
	});

	it('disables the select', () => {
		render(Select, { props: { disabled: true, children: optionsSnippet() } });
		const select = screen.getByRole('combobox');
		expect(select).toBeDisabled();
	});

	it('calls onchange when value changes', async () => {
		const onchange = vi.fn();
		render(Select, { props: { onchange, children: optionsSnippet() } });
		const select = screen.getByRole('combobox');
		await fireEvent.change(select, { target: { value: 'b' } });
		expect(onchange).toHaveBeenCalled();
	});

	it('applies custom class', () => {
		const { container } = render(Select, {
			props: { class: 'my-select', children: optionsSnippet() }
		});
		expect(container.querySelector('div.my-select')).not.toBeNull();
	});
});

// =============================================================================
// Tabs
// =============================================================================
describe('Tabs', () => {
	const baseTabs = [
		{ value: 'one', label: 'Tab One' },
		{ value: 'two', label: 'Tab Two' },
		{ value: 'three', label: 'Tab Three' }
	];

	it('renders all tab buttons', () => {
		render(Tabs, { props: { tabs: baseTabs, value: 'one' } });
		expect(screen.getAllByRole('tab')).toHaveLength(3);
		expect(screen.getByText('Tab One')).toBeInTheDocument();
		expect(screen.getByText('Tab Two')).toBeInTheDocument();
		expect(screen.getByText('Tab Three')).toBeInTheDocument();
	});

	it('marks the active tab with aria-selected', () => {
		render(Tabs, { props: { tabs: baseTabs, value: 'two' } });
		const tabs = screen.getAllByRole('tab');
		expect(tabs[0]).toHaveAttribute('aria-selected', 'false');
		expect(tabs[1]).toHaveAttribute('aria-selected', 'true');
		expect(tabs[2]).toHaveAttribute('aria-selected', 'false');
	});

	it('calls onchange when a tab is clicked', async () => {
		const onchange = vi.fn();
		render(Tabs, { props: { tabs: baseTabs, value: 'one', onchange } });
		await fireEvent.click(screen.getByText('Tab Three'));
		expect(onchange).toHaveBeenCalledWith('three');
	});

	it('renders tab icon when provided', () => {
		const tabsWithIcon = [{ value: 'x', label: 'With Icon', icon: '/logos/sonarr.png' }];
		const { container } = render(Tabs, { props: { tabs: tabsWithIcon, value: 'x' } });
		const img = container.querySelector('img');
		expect(img).not.toBeNull();
		expect(img!.getAttribute('src')).toBe('/logos/sonarr.png');
	});

	it('has tablist role on container', () => {
		render(Tabs, { props: { tabs: baseTabs, value: 'one' } });
		expect(screen.getByRole('tablist')).toBeInTheDocument();
	});

	it('applies custom class to container', () => {
		render(Tabs, { props: { tabs: baseTabs, value: 'one', class: 'extra-tabs' } });
		const tablist = screen.getByRole('tablist');
		expect(tablist.className).toContain('extra-tabs');
	});
});

// =============================================================================
// Skeleton
// =============================================================================
describe('Skeleton', () => {
	it('renders one row by default', () => {
		const { container } = render(Skeleton);
		const pulses = container.querySelectorAll('.animate-pulse');
		expect(pulses).toHaveLength(1);
	});

	it('renders the specified number of rows', () => {
		const { container } = render(Skeleton, { props: { rows: 4 } });
		const pulses = container.querySelectorAll('.animate-pulse');
		expect(pulses).toHaveLength(4);
	});

	it('applies default height class', () => {
		const { container } = render(Skeleton);
		expect(container.querySelector('.h-10')).not.toBeNull();
	});

	it('applies custom height class', () => {
		const { container } = render(Skeleton, { props: { height: 'h-20' } });
		expect(container.querySelector('.h-20')).not.toBeNull();
	});

	it('applies custom class', () => {
		const { container } = render(Skeleton, { props: { class: 'my-skeleton' } });
		expect(container.querySelector('.my-skeleton')).not.toBeNull();
	});
});

// =============================================================================
// EmptyState
// =============================================================================
describe('EmptyState', () => {
	it('renders title', () => {
		render(EmptyState, { props: { title: 'Nothing here' } });
		expect(screen.getByText('Nothing here')).toBeInTheDocument();
	});

	it('renders description when provided', () => {
		render(EmptyState, {
			props: { title: 'Empty', description: 'Add something to get started.' }
		});
		expect(screen.getByText('Add something to get started.')).toBeInTheDocument();
	});

	it('does not render description when not provided', () => {
		const { container } = render(EmptyState, { props: { title: 'Empty' } });
		expect(container.querySelector('p.text-muted-foreground')).toBeNull();
	});

	it('renders actions snippet when provided', () => {
		render(EmptyState, {
			props: { title: 'Empty', actions: textSnippet('Add Item') }
		});
		expect(screen.getByText('Add Item')).toBeInTheDocument();
	});

	it('applies custom class', () => {
		const { container } = render(EmptyState, {
			props: { title: 'Empty', class: 'my-empty' }
		});
		// Card wraps contents in a div — find the one with our class
		expect(container.querySelector('.my-empty')).not.toBeNull();
	});
});

// =============================================================================
// Separator
// =============================================================================
describe('Separator', () => {
	it('renders with separator role', () => {
		render(Separator);
		expect(screen.getByRole('separator')).toBeInTheDocument();
	});

	it('defaults to horizontal orientation', () => {
		render(Separator);
		const sep = screen.getByRole('separator');
		expect(sep).toHaveAttribute('aria-orientation', 'horizontal');
		expect(sep.className).toContain('h-px');
		expect(sep.className).toContain('w-full');
	});

	it('renders vertical orientation', () => {
		render(Separator, { props: { orientation: 'vertical' } });
		const sep = screen.getByRole('separator');
		expect(sep).toHaveAttribute('aria-orientation', 'vertical');
		expect(sep.className).toContain('w-px');
		expect(sep.className).toContain('h-full');
	});

	it('applies custom class', () => {
		render(Separator, { props: { class: 'my-sep' } });
		const sep = screen.getByRole('separator');
		expect(sep.className).toContain('my-sep');
	});
});

// =============================================================================
// PageHeader
// =============================================================================
describe('PageHeader', () => {
	it('renders title as h1', () => {
		render(PageHeader, { props: { title: 'Dashboard' } });
		const h1 = screen.getByRole('heading', { level: 1 });
		expect(h1).toHaveTextContent('Dashboard');
	});

	it('renders description when provided', () => {
		render(PageHeader, {
			props: { title: 'Settings', description: 'Manage your preferences' }
		});
		expect(screen.getByText('Manage your preferences')).toBeInTheDocument();
	});

	it('does not render description when not provided', () => {
		const { container } = render(PageHeader, { props: { title: 'Settings' } });
		expect(container.querySelector('p.text-muted-foreground')).toBeNull();
	});

	it('renders actions snippet when provided', () => {
		render(PageHeader, {
			props: { title: 'Page', actions: textSnippet('Save') }
		});
		expect(screen.getByText('Save')).toBeInTheDocument();
	});

	it('applies custom class', () => {
		const { container } = render(PageHeader, {
			props: { title: 'Page', class: 'my-header' }
		});
		expect(container.querySelector('.my-header')).not.toBeNull();
	});
});

// =============================================================================
// DataTable
// =============================================================================
describe('DataTable', () => {
	const columns = [
		{ key: 'name', header: 'Name', sortable: true },
		{ key: 'status', header: 'Status' }
	];
	const data = [
		{ name: 'Alice', status: 'Active' },
		{ name: 'Bob', status: 'Inactive' }
	];
	const rowSnippet = createRawSnippet(() => ({
		render: () => `<tr><td>row</td></tr>`
	}));

	it('renders a table element', () => {
		render(DataTable, { props: { columns, data, row: rowSnippet } });
		expect(screen.getByRole('table')).toBeInTheDocument();
	});

	it('renders column headers', () => {
		render(DataTable, { props: { columns, data, row: rowSnippet } });
		expect(screen.getByText('Name')).toBeInTheDocument();
		expect(screen.getByText('Status')).toBeInTheDocument();
	});

	it('applies custom class to outer container', () => {
		const { container } = render(DataTable, {
			props: { class: 'my-table', columns, data, row: rowSnippet }
		});
		expect(container.querySelector('.my-table')).not.toBeNull();
	});

	it('renders search input when searchable', () => {
		render(DataTable, { props: { columns, data, row: rowSnippet, searchable: true } });
		expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument();
	});
});
