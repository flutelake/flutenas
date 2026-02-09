<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import NotFound from '../components/errors/NotFound.svelte';
	import Maintenance from '../components/errors/Maintenance.svelte';
	import ServerError from '../components/errors/ServerError.svelte';

	const pages = {
		400: Maintenance,
		404: NotFound,
		500: ServerError
	};

	const status = +$page.status;
	const index = Object.keys(pages)
		.map((x) => +x)
		.reduce((p, c) => (p < status ? c : p));
	const component = pages[index as keyof typeof pages];

	import MetaTag from '../components/MetaTag.svelte';

	const path: string = `/errors/${index}`;
	const description: string = `${index} - FluteNAS Web Console`;
	const title: string = ` FluteNAS Web Console - ${index} page`;
	const subtitle: string = `${index} page`;
</script>

<MetaTag {path} {description} {title} {subtitle} />

<svelte:component this={component}></svelte:component>
