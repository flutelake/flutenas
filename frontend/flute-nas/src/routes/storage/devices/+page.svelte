<script lang="ts">
	import { onMount } from 'svelte';
	import { Breadcrumb, BreadcrumbItem, Button, Checkbox, Drawer, Heading, Input, Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Toolbar, ToolbarButton, Badge, Hr } from 'flowbite-svelte';
	import { CogSolid, DotsVerticalOutline, EditOutline, ExclamationCircleSolid, TrashBinSolid, RefreshOutline } from 'flowbite-svelte-icons';
	import type { ComponentType } from 'svelte';
	import { sineIn } from 'svelte/easing';
	import MetaTag from '../../../components/MetaTag.svelte';
	import { FluteAPI } from '$lib/api';
	import { DiskDevice } from '$lib/model';
	import SetMountPointModal from '../../../components/storage/SetMountPointModal.svelte';
	export let node : string = 'localhost';
	import { CurrentHostIP } from '$lib/vars';

	let hidden: boolean = true; // modal control
	let loading: boolean = false;

	const toggle = (component: ComponentType) => {
		hidden = !hidden;
	};

	const path: string = '/storage/devices';
  	const description: string = 'storage devices - FluteNAS Web Console';
	const title: string = 'FluteNAS Web Console - Storage Devices';
	const subtitle: string = 'Storage Devices';
	let transitionParams = {
		x: 320,
		duration: 200,
		easing: sineIn
	};

	let devices :DiskDevice[] = [];
	$: refreshList($CurrentHostIP)
	function refreshList(ip :string) {
		// console.log('refreshList, ip: ', ip)
		if (loading) {
			// 防重复点击
			return
		}
		if (ip == "") {
			ip = "127.0.0.1"
		}
		loading = true;
		console.log(ip)
		const api = new FluteAPI();
        api.post("/v1/disk/list", {"HostIP": ip}).then(resp => {
			devices = DiskDevice.UmarshalArray(resp.data.Devices);
			loading = false;
        }).catch(err => {
            console.log(err)
			loading = false;
        })
	}
 	// onMount(() => {
	// 	refreshList($CurrentHostIP)
	// })

	let openSetMountPointModal :boolean = false;
	let currentDevice :DiskDevice;
	function toggleSetMountPointModal(d :DiskDevice) {
		console.log('open SetMountPoint Dialog')
		currentDevice = d;
		openSetMountPointModal = !openSetMountPointModal
	}
</script>

<MetaTag {path} {description} {title} {subtitle} />

<main class="relative h-full w-full overflow-y-auto bg-white dark:bg-gray-800">
	<div class="p-4">
		<Breadcrumb class="mb-5">
			<BreadcrumbItem href="/" home>Home</BreadcrumbItem>
			<BreadcrumbItem>Storage</BreadcrumbItem>
			<BreadcrumbItem>Devices</BreadcrumbItem>
		</Breadcrumb>
		<Heading tag="h1" class="text-xl font-semibold text-gray-900 dark:text-white sm:text-2xl">
			Disk Devices
		</Heading>
		<Toolbar embedded class="w-full py-4 text-gray-500 dark:text-gray-400">
			<div slot="end" class="space-x-2">
				<!-- on:click={() => toggle("")} -->
				<Button pill={true} class="!p-2 text-white bg-gradient-to-br from-purple-600 to-blue-500 hover:bg-gradient-to-bl focus:ring-blue-300 dark:focus:ring-blue-800 shadow-blue-500/50 dark:shadow-blue-800/80" on:click={() => {refreshList($CurrentHostIP)}} >
					{#if loading}
					<RefreshOutline class="w-6 h-6 spin-fast"/>&nbsp; Loading... Please wait
					{:else}
					<RefreshOutline class="w-6 h-6"/>
					{/if}
				</Button>
				<!-- <Button class="whitespace-nowrap" >Add new product</Button> -->
			</div>
		</Toolbar>
	</div>
	<Table>
		<TableHead class="border-y border-gray-200 bg-gray-100 dark:border-gray-700">
			<TableHeadCell class="w-4 p-4"><Checkbox /></TableHeadCell>
			{#each ['Device Name', 'Size', 'Serial', 'MountPoint', 'Vendor', 'FileSystem', 'Labels', 'Actions'] as title}
				<TableHeadCell class="ps-4 font-normal">{title}</TableHeadCell>
			{/each}
		</TableHead>
		<TableBody>
			{#each devices as d}
				<TableBodyRow class="text-base">
					<TableBodyCell class="w-4 p-4"><Checkbox /></TableBodyCell>
					<TableBodyCell class="max-w-sm overflow-hidden truncate p-4 text-base font-normal text-gray-500 dark:text-gray-400 xl:max-w-xs">
						<div class="text-sm font-normal text-gray-500 dark:text-gray-400">
							<div class="text-base font-semibold text-gray-900 dark:text-white">
								{d.Name.replace('/dev/', '')}
							</div>
							{#if d.IsSystemDisk }
							<div class="text-sm font-normal text-gray-500 dark:text-gray-400">
								{ d.IsSystemDisk ? 'BootDisk' : '' }
							</div>
							{/if}
						</div>
					</TableBodyCell>
					<TableBodyCell class="p-4">{d.Size}</TableBodyCell>
					<TableBodyCell class="p-4">{d.Serial}</TableBodyCell>
					<TableBodyCell
						class="max-w-sm overflow-hidden truncate p-4 text-base font-normal text-gray-500 dark:text-gray-400 xl:max-w-xs"
						>
						<div class="text-sm font-normal text-gray-500 dark:text-gray-400">
							<div class="text-base font-semibold text-gray-900 dark:text-white">
								{d.MountPoint}
							</div>
							{#if d.SpecMountPoint }
							<!-- todo Pending 可以加一个问号的小图标解释其含义 -->
							<div class="text-sm font-normal text-gray-500 dark:text-gray-400">
								Pending: { d.SpecMountPoint }
							</div>
							{/if}
						</div>
					</TableBodyCell>
					<TableBodyCell class="p-4">{d.Vendor}</TableBodyCell>
					<TableBodyCell class="p-4">{d.FsType}</TableBodyCell>
					<!-- <TableBodyCell class="p-4">{d.PartUUID == '' ? 'No' : 'Yes'}</TableBodyCell> -->
					<TableBodyCell class="p-4">
						{#if d.Labels}
							{#each d.Labels as l}
								<Badge color="indigo">{l}</Badge>
							{/each}
						{/if}
					</TableBodyCell>
					<TableBodyCell class="space-x-2">
						{#if d.IsSystemDisk}
							<Button  disabled size="sm" class="gap-2 px-3">
								<EditOutline size="sm" /> mkfs
							</Button>
							<Button disabled color="red" size="sm" class="gap-2 px-3" on:click={() => toggleSetMountPointModal(d)}>
								<TrashBinSolid size="sm" /> setMountPoint
							</Button>
						{:else}
							<Button size="sm" class="gap-2 px-3">
								<EditOutline size="sm" /> mkfs
							</Button>
							<Button color="red" size="sm" class="gap-2 px-3" on:click={() => toggleSetMountPointModal(d)}>
								<TrashBinSolid size="sm" /> setMountPoint
							</Button>
						{/if}
						
					</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>

	<Hr />
	<div class="p-4">
		<Heading
		tag="h1"
		size="xl"
		class="mb-3 text-3xl font-bold text-gray-900 dark:text-white sm:text-4xl sm:leading-none sm:tracking-tight"
		>
		FAQ
		</Heading>
		<p class="mb-6 text-lg font-normal text-gray-500 dark:text-gray-400 sm:text-xl">
			frequently asked questions
		</p>
	
		<div class="lg:columns-1 gap-8 space-y-10">
			<div class="space-y-4">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">
					What different set mount-point in page with mount manually?
				</h3>
				<p class="text-gray-600 dark:text-gray-400">
					Set the mount-point on the page, and flutenas will maintain the mount point after system restart, or after manual modification, flutenas will automatically restore it.
				</p>
				<p class="text-gray-600 dark:text-gray-400">
					If you do not want flutenas take over mount-point, you should set mount-point with empty string.
				</p>
			</div>

			<div class="space-y-4">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">
					Why set mount-point button is disabled?
				</h3>
				<p class="text-gray-600 dark:text-gray-400">
					If there are partitions on the current disk device, the set mount-point button will be disabled, or the disk is operating system disk.
				</p>
			</div>

			<div class="space-y-4">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">
					My hard disk is 4T, but capacity in table only show 3.64TiB, why?
				</h3>
				<p class="text-gray-600 dark:text-gray-400">
					Usually the advertised capacity on hard disk products is converted based on 1000, but in flutenas it is converted based on 1024, so the hard disk capacity will be slightly smaller than the advertised capacity.
					If there is no special agreement, the unit `TiB` represents the conversion basis of 1024, and the unit `TB` represents the conversion basis of 1000.
				</p>
			</div>
		</div>
	</div>
</main>

<SetMountPointModal bind:open={openSetMountPointModal} bind:node={node} bind:disk={currentDevice} on:refesh_list_msg={()=>refreshList($CurrentHostIP)}></SetMountPointModal>