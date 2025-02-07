<script lang="ts">
	import { avatarPath } from '../variables'
	import { Avatar, Dropdown, DropdownDivider, DropdownHeader, DropdownItem } from 'flowbite-svelte';

	export let name: string = ''; // "Neil Sims",
	export let avatar: string = ''; // "neil-sims.png",
	export let email: string = ''; // "neil.sims@flowbite.com",
	import { FluteAPI } from '$lib/api';
	import { goto } from '$app/navigation';
	function handleSignOut() {
		const api = new FluteAPI();
        api.post("/v1/logout", {}).then(resp => {
			console.log('Sign out');
			goto('/login')
        }).catch(err => {
            console.log(err)
        })
	}
</script>

<button class="ms-3 rounded-full ring-gray-400 focus:ring-4 dark:ring-gray-600">
	<Avatar size="sm" src={avatarPath(avatar)} tabindex="0" />
</button>
<Dropdown placement="bottom-end">
	<DropdownHeader>
		<span class="block text-sm">{name}</span>
		<span class="block truncate text-sm font-medium">{email}</span>
	</DropdownHeader>
	<!-- <DropdownItem>Dashboard</DropdownItem>
	<DropdownItem>Settings</DropdownItem>
	<DropdownItem>Earnings</DropdownItem>
	<DropdownDivider /> -->
	<DropdownItem on:click={() => handleSignOut()}>Sign out</DropdownItem>
</Dropdown>

<!--
@component
[Go to docs](https://flowbite-svelte-admin-dashboard.vercel.app/)
## Props
@prop export let id: number = 0;
@prop export let name: string = '';
@prop export let avatar: string = '';
@prop export let email: string = '';
@prop export let biography: string = '';
@prop export let position: string = '';
@prop export let country: string = '';
@prop export let status: string = '';
-->
