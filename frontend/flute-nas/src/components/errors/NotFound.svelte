<script lang="ts">
	import { Button, P } from 'flowbite-svelte';
	export let title = 'Page not found';
	export let description =
		'Oops! Looks like you followed a bad link. If you think this is a problem with us, please	tell us.';
	export let image = {
		src: 'https://flowbite-admin-dashboard.vercel.app/images/illustrations/404.svg',
		alt: 'astronaut'
	};
	export let btnTitle = 'Go back home';
	export let btnHref = '/overview';
	export let mainClass = 'bg-gray-50 dark:bg-gray-900';
	export let mainDivClass =
		'flex flex-col items-center justify-center px-6 pt-8 mx-auto md:h-screen pt:mt-0 dark:bg-gray-900';
	export let h1Class =
		'mb-3 text-2xl font-bold leading-tight text-gray-900 sm:text-4xl lg:text-5xl dark:text-white';
	export let imgDiv = 'block md:max-w-lg';
	export let div2Class = 'text-center xl:max-w-4xl';
</script>

<main class={mainClass}>
	<div class={mainDivClass}>
		<div class={imgDiv}>
			<img src={image.src} alt={image.alt} />
		</div>
		<div class={div2Class}>
			<h1 class={h1Class}>
				{title}
			</h1>
			<P class="mb-5 text-base font-normal text-gray-500 dark:text-gray-400 md:text-lg">
				{description}
			</P>
			<Button href={btnHref}>
				<svg
					class="-ml-1 mr-2 h-5 w-5"
					fill="currentColor"
					viewBox="0 0 20 20"
					xmlns="http://www.w3.org/2000/svg"
					><path
						fill-rule="evenodd"
						d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
						clip-rule="evenodd"
					/></svg
				>
				{btnTitle}
			</Button>
		</div>
	</div>
</main>

<!--
@component
[Go to docs](https://flowbite-svelte-admin-dashboard.vercel.app/)
## Props
@prop export let title = 'Page not found';
@prop export let description =
		'Oops! Looks like you followed a bad link. If you think this is a problem with us, please	tell us.';
@prop export let image = {
		src: imagesPath('404.svg', 'illustrations'),
		alt: 'astronaut'
	};
@prop export let btnTitle = 'Go back home';
@prop export let btnHref = '/';
@prop export let mainClass = 'bg-gray-50 dark:bg-gray-900';
@prop export let mainDivClass =
		'flex flex-col items-center justify-center px-6 pt-8 mx-auto md:h-screen pt:mt-0 dark:bg-gray-900';
@prop export let h1Class =
		'mb-3 text-2xl font-bold leading-tight text-gray-900 sm:text-4xl lg:text-5xl dark:text-white';
@prop export let imgDiv = 'block md:max-w-lg';
@prop export let div2Class = 'text-center xl:max-w-4xl';
-->
