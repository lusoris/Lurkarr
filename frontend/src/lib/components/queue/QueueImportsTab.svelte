<script lang="ts">
	import Card from '$lib/components/ui/Card.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import DataTable, { type Column } from '$lib/components/ui/DataTable.svelte';
	import * as T from '$lib/components/ui/table';
	import Badge from '$lib/components/ui/Badge.svelte';
	import type { AutoImportLog } from '$lib/types';

	interface Props {
		app: string;
		imports: AutoImportLog[];
		loading: boolean;
	}

	let { app, imports, loading }: Props = $props();

	const importLogColumns: Column<AutoImportLog>[] = [
		{ key: 'media_title', header: 'Media', sortable: true },
		{ key: 'action', header: 'Action', sortable: true },
		{ key: 'reason', header: 'Reason' },
		{ key: 'created_at', header: 'Date', sortable: true }
	];
</script>

{#if loading}
	<Skeleton rows={4} height="h-10" />
{:else if imports.length === 0}
	<EmptyState title="No import entries" description="Auto-imported items will appear here." />
{:else}
	<DataTable data={imports} columns={importLogColumns} searchable pageSize={50} noun="imports">
		{#snippet row(entry)}
			<T.Row>
				<T.Cell class="text-foreground">{entry.media_title}</T.Cell>
				<T.Cell><Badge variant="info">{entry.action}</Badge></T.Cell>
				<T.Cell class="text-muted-foreground max-w-xs truncate">{entry.reason}</T.Cell>
				<T.Cell class="text-muted-foreground text-xs">{new Date(entry.created_at).toLocaleString()}</T.Cell>
			</T.Row>
		{/snippet}
	</DataTable>
{/if}
