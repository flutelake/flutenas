<script lang="ts">
    import MetaTag from '../../../components/MetaTag.svelte';
    import { Avatar, Breadcrumb, BreadcrumbItem, Button, Checkbox, Heading, Indicator } from 'flowbite-svelte';

	const path: string = '/overview';
    const description: string = 'Overview - flute nas console';
	const metaTitle: string = 'FluteNAS Web Console - Overview';
    const subtitle: string = 'overview';
</script>

<MetaTag {path} {description} title={metaTitle} {subtitle} />

<main class="relative h-full w-full overflow-y-auto bg-white dark:bg-gray-800">
	<div class="p-4">
		<Breadcrumb class="mb-5">
			<BreadcrumbItem home>Home</BreadcrumbItem>
			<BreadcrumbItem href="/overview">Overview</BreadcrumbItem>
		</Breadcrumb>
		<Heading tag="h1" class="text-xl font-semibold text-gray-900 dark:text-white sm:text-2xl">
			Overview
		</Heading>
	</div>
</main>

<!-- Modals -->

<!-- <User bind:open={openUser} data={current_user} />
<Delete bind:open={openDelete} /> -->
