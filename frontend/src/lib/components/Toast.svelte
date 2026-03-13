<script lang="ts">
	import { Toaster } from 'svelte-sonner';
</script>

<Toaster
	position="bottom-right"
	theme="dark"
	richColors
	closeButton
	toastOptions={{
		class: 'border-border bg-card text-card-foreground'
	}}
/>
