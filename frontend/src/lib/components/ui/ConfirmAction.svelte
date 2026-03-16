<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as AlertDialog from './alert-dialog';

	interface Props {
		/** Whether the confirmation prompt is currently shown. */
		active: boolean;
		/** Label for the confirm prompt. */
		message?: string;
		/** Optional title for the dialog header. */
		title?: string;
		/** Callback when the user confirms. */
		onconfirm: () => void;
		/** Callback when the user cancels (resets active state). */
		oncancel: () => void;
		/** Content to show when not confirming (the trigger button). */
		children: Snippet;
	}

	let { active, message = 'Are you sure?', title = 'Confirm Action', onconfirm, oncancel, children }: Props = $props();
</script>

{@render children()}

<AlertDialog.Root open={active} onOpenChange={(open) => { if (!open) oncancel(); }}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>{title}</AlertDialog.Title>
			<AlertDialog.Description>{message}</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={oncancel}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action class="bg-destructive text-white hover:bg-destructive/90" onclick={onconfirm}>Confirm</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
