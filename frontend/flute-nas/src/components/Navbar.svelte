<script lang="ts">
	import Notifications from './dashboard/NotificationList.svelte';
	import AppsMenu from './widgets/AppsMenu.svelte';
	import UserMenu from './widgets/UserMenu.svelte';
	import {
		DarkMode,
		Dropdown,
		DropdownItem,
		NavBrand,
		NavHamburger,
		NavLi,
		NavUl,
		Navbar,
		Search,
		Button
	} from 'flowbite-svelte';
	import { ChevronDownOutline } from 'flowbite-svelte-icons';
	import '../app.css';
	import Users from '../data/users.json';
	import { type Host } from '$lib/interface'
	import { onMount } from 'svelte';
	import { FluteAPI } from '$lib/api';
	import { CurrentHostIP, CurrentHostIPKey } from '$lib/vars'

	export let fluid = true;
	export let drawerHidden = false;
	let hosts: Host[] = [];
	let selectHost :Host;
	// 节点下拉选择框选中的样式
	let activeClass = 'text-green-500 dark:text-green-300 hover:text-green-700 dark:hover:text-green-500';
	
	onMount(() => {
		const api = new FluteAPI()
		api.post("/v1/host/list", {}).then(resp => {
			hosts = resp.data.Hosts
			let existHostIP = sessionStorage.getItem(CurrentHostIPKey)
			if (existHostIP) {
				// CurrentHostIP.set(existHostIP)
				selectHost = hosts.find(host => host.HostIP === existHostIP) || hosts[0]
			} else {
				selectHost = hosts[0]
			}
			// console.log(hosts)
		}).catch(err => {
			console.log(err)
		})
		// 读取当前目录下的文件
		// readDir(dirPath)
	})
	let list = true;

	function handleSelectHost(e: any) {
		e.preventDefault();
		console.log("select host: ", e.target.innerText)
		const hostname = e.target.innerText
		let sh = hosts.find(host => host.Hostname === hostname)
		if (sh) {
			selectHost = sh
			// 更新全局变量
			CurrentHostIP.set(selectHost.HostIP.toString())
			sessionStorage.setItem(CurrentHostIPKey, selectHost.HostIP)
		}
		console.log(selectHost)
	}
</script>

<Navbar {fluid} class="text-black" color="default" let:NavContainer>
	<NavContainer class="mb-px mt-px px-1" {fluid}>
		<NavHamburger
			onClick={() => (drawerHidden = !drawerHidden)}
			class="m-0 me-3 md:block lg:hidden"
		/>
		<NavBrand href="/" class={list ? 'w-40' : 'lg:w-60'}>
			<img
				src="/images/flowbite-svelte-icon-logo.svg"
				class="me-2.5 h-6 sm:h-8"
				alt="Flowbite Logo"
			/>
			<span
				class="ml-px self-center whitespace-nowrap text-xl font-semibold dark:text-white sm:text-2xl"
			>
				FluteNAS
			</span>
		</NavBrand>
		<div class="hidden lg:block lg:ps-3">
			{#if list}
				<NavUl class="ml-2" activeUrl="/" activeClass="text-primary-600 dark:text-primary-500">
					<NavLi href="/">Home</NavLi>
					<NavLi href="/terminal">Terminal</NavLi>
					<NavLi href="/storage/devices">Storage</NavLi>
					<NavLi href="/filestation">Files</NavLi>
					<NavLi href="#top">Settings</NavLi>

					<NavLi class="cursor-pointer primary">
						{selectHost?.Hostname}
						<ChevronDownOutline  class="ms-0 inline" />
					</NavLi>
					<Dropdown class="z-20 w-44" >
						{#each hosts as host}
							{#if host.HostIP == selectHost.HostIP}
								<DropdownItem class={activeClass}>{host.Hostname}</DropdownItem>
							{:else}
								<DropdownItem href="{host.HostIP}" on:click={handleSelectHost}>{host.Hostname}</DropdownItem>
							{/if}
							<!-- <a href="#" on:click={handleSelectHost}>{host.hostname}</a> -->
							<!-- <DropdownItem href="{host.host_ip}" on:click={handleSelectHost}>{host.hostname}</DropdownItem> -->
						{/each}
					</Dropdown>
				</NavUl>
			{:else}
				<form>
					<Search size="md" class="mt-1 w-96 border focus:outline-none" />
				</form>
			{/if}
		</div>
		<div class="ms-auto flex items-center text-gray-500 dark:text-gray-400 sm:order-2">
			<Notifications />
			<AppsMenu />
			<DarkMode />
			<UserMenu {...Users[0]} />
		</div>
	</NavContainer>
</Navbar>
