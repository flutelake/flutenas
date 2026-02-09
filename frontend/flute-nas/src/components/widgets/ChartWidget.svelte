<script lang="ts">
	import type { ApexOptions } from 'apexcharts';
	import { Card, Chart, Heading } from 'flowbite-svelte';
	import LastRange from './LastRange.svelte';
	import More from './More.svelte';

	export let title: string = '';
	export let subtitle: string = '';
	export let chartOptions: ApexOptions;
	export let timeslot: string | undefined = undefined;
	export let timeslots: Record<string, number> | undefined = undefined;
</script>

<Card size="xl" class="w-full max-w-none 2xl:col-span-2">
	<div class="mb-4 flex items-center justify-between">
		<div class="flex-shrink-0">
			<Heading tag="h3" class="text-2xl">{title}</Heading>
			<p class="text-base font-light text-gray-500 dark:text-gray-400">{subtitle}</p>
		</div>
	</div>

	<Chart options={chartOptions}></Chart>
	<div
		class="mt-4 flex items-center justify-between border-t border-gray-200 pt-3 sm:pt-6 dark:border-gray-700"
	>
		{#if timeslots}
			<LastRange bind:timeslot {timeslots} />
		{/if}
	</div>
</Card>

<!--
@component
[Go to docs](https://flowbite-svelte-admin-dashboard.vercel.app/)
## Props
@prop export let title: string = '';
@prop export let subtitle: string = '';
@prop export let chartOptions: ApexOptions;
-->
