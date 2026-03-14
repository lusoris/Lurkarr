<script lang="ts" module>
	export interface Column<T = any> {
		key: string;
		header: string;
		sortable?: boolean;
		/** Extract a sortable/filterable value from the row. Defaults to (row as any)[key]. */
		accessor?: (row: T) => string | number;
		class?: string;
		headerClass?: string;
	}
</script>

<script lang="ts" generics="T">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { ArrowUpDown, ArrowUp, ArrowDown, ChevronLeft, ChevronRight } from 'lucide-svelte';
	import * as Table from './table';

	interface Props {
		data: T[];
		columns: Column<T>[];
		/** Snippet to render each row. Receives (item, index). */
		row: Snippet<[T, number]>;
		/** Enable text search filtering. */
		searchable?: boolean;
		searchPlaceholder?: string;
		/** External search value (two-way bindable). */
		search?: string;
		/** Rows per page (0 = no pagination). */
		pageSize?: number;
		/** External page value (two-way bindable, 1-based). */
		page?: number;
		/** If provided, used for total count display (for server-side pagination). */
		totalItems?: number;
		/** Label shown next to pagination count. */
		noun?: string;
		class?: string;
		/** Extra content above the table (e.g. filters). */
		toolbar?: Snippet;
	}

	let {
		data,
		columns,
		row,
		searchable = false,
		searchPlaceholder = 'Search...',
		search = $bindable(''),
		pageSize = 0,
		page = $bindable(1),
		totalItems,
		noun = 'items',
		class: className = '',
		toolbar
	}: Props = $props();

	// --- Sorting ---
	let sortKey = $state('');
	let sortDir = $state<'asc' | 'desc'>('asc');

	function toggleSort(key: string) {
		if (sortKey === key) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortKey = key;
			sortDir = 'asc';
		}
		page = 1;
	}

	function getValue(item: T, col: Column<T>): string | number {
		if (col.accessor) return col.accessor(item);
		return (item as Record<string, any>)[col.key] ?? '';
	}

	// --- Filtered + sorted data ---
	const processed = $derived.by(() => {
		let result = data;

		// Client-side search filter
		if (searchable && search.trim()) {
			const q = search.trim().toLowerCase();
			result = result.filter((item) =>
				columns.some((col) => String(getValue(item, col)).toLowerCase().includes(q))
			);
		}

		// Sort
		if (sortKey) {
			const col = columns.find((c) => c.key === sortKey);
			if (col) {
				result = [...result].sort((a, b) => {
					const va = getValue(a, col);
					const vb = getValue(b, col);
					const cmp = typeof va === 'number' && typeof vb === 'number'
						? va - vb
						: String(va).localeCompare(String(vb));
					return sortDir === 'asc' ? cmp : -cmp;
				});
			}
		}

		return result;
	});

	// --- Pagination ---
	const isServerPaged = $derived(totalItems !== undefined);
	const totalCount = $derived(isServerPaged ? totalItems! : processed.length);
	const totalPages = $derived(pageSize > 0 ? Math.max(1, Math.ceil(totalCount / pageSize)) : 1);
	const paged = $derived(
		pageSize > 0 && !isServerPaged
			? processed.slice((page - 1) * pageSize, page * pageSize)
			: processed
	);

	// Clamp page if out of range
	$effect(() => {
		if (page > totalPages) page = totalPages;
		if (page < 1) page = 1;
	});
</script>

<div class={cn('space-y-3', className)}>
	{#if searchable || toolbar}
		<div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
			{#if searchable}
				<input
					type="text"
					placeholder={searchPlaceholder}
					bind:value={search}
					class={cn(
						'border-input bg-background placeholder:text-muted-foreground flex h-9 w-full sm:max-w-xs rounded-md border px-3 py-1 text-sm shadow-xs outline-none transition-[color,box-shadow]',
						'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]'
					)}
				/>
			{/if}
			{#if toolbar}
				{@render toolbar()}
			{/if}
		</div>
	{/if}

	<div class="rounded-xl border border-border overflow-hidden">
		<Table.Root>
			<Table.Header>
				<Table.Row class="bg-muted/50">
					{#each columns as col}
						<Table.Head class={cn(col.headerClass)}>
							{#if col.sortable}
								<button
									type="button"
									class="inline-flex items-center gap-1 hover:text-foreground transition-colors"
									onclick={() => toggleSort(col.key)}
								>
									{col.header}
									{#if sortKey === col.key}
										{#if sortDir === 'asc'}
											<ArrowUp class="h-3 w-3" />
										{:else}
											<ArrowDown class="h-3 w-3" />
										{/if}
									{:else}
										<ArrowUpDown class="h-3 w-3 opacity-30" />
									{/if}
								</button>
							{:else}
								{col.header}
							{/if}
						</Table.Head>
					{/each}
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each paged as item, i}
					{@render row(item, (page - 1) * (pageSize || 1) + i)}
				{:else}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="text-center text-muted-foreground py-8">
							{searchable && search ? 'No results match your search.' : 'No data.'}
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>

	{#if pageSize > 0 && totalCount > 0}
		<div class="flex items-center justify-between text-sm text-muted-foreground">
			<span>{totalCount.toLocaleString()} {noun}</span>
			{#if totalPages > 1}
				<div class="flex items-center gap-2">
					<button type="button" disabled={page <= 1} onclick={() => page--}
						class="p-1 rounded hover:bg-muted disabled:opacity-30 disabled:cursor-not-allowed">
						<ChevronLeft class="h-4 w-4" />
					</button>
					<span>Page {page} of {totalPages}</span>
					<button type="button" disabled={page >= totalPages} onclick={() => page++}
						class="p-1 rounded hover:bg-muted disabled:opacity-30 disabled:cursor-not-allowed">
						<ChevronRight class="h-4 w-4" />
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>
