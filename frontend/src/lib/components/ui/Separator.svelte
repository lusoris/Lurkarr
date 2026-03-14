<script lang="ts">
	import { Separator as ShadcnSeparator } from './separator';

	interface Props {
		orientation?: 'horizontal' | 'vertical';
		class?: string;
	}

	let { orientation = 'horizontal', class: className = '' }: Props = $props();
</script>

<ShadcnSeparator {orientation} class={className} />
