<script lang="ts">
	import { onMount } from 'svelte';
	import { Breadcrumb, BreadcrumbItem, Button, Checkbox, Drawer, Heading, Input, Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Toolbar, ToolbarButton, Badge, Hr } from 'flowbite-svelte';
	import { CogSolid, DotsVerticalOutline, EditOutline, ExclamationCircleSolid, TrashBinSolid, RefreshOutline } from 'flowbite-svelte-icons';
	import type { ComponentType } from 'svelte';
	import { sineIn } from 'svelte/easing';
	import MetaTag from '../../../components/MetaTag.svelte';
	import { FluteAPI } from '$lib/api';
	import { DiskDevice } from '$lib/model';

	let hidden: boolean = true; // modal control

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
 	onMount(() => {
		console.log(111)
		const api = new FluteAPI();
        api.post("/v1/disk/list", {}).then(resp => {
			devices = DiskDevice.UmarshalArray(resp.data.Devices);
        }).catch(err => {
            console.log(err)
        })
	})
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
				<Button pill={true} class="!p-2 text-white bg-gradient-to-br from-purple-600 to-blue-500 hover:bg-gradient-to-bl focus:ring-blue-300 dark:focus:ring-blue-800 shadow-blue-500/50 dark:shadow-blue-800/80"><RefreshOutline class="w-6 h-6" /></Button>
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
					<TableBodyCell class="flex items-center space-x-6 whitespace-nowrap p-4">
						<div class="text-sm font-normal text-gray-500 dark:text-gray-400">
							<div class="text-base font-semibold text-gray-900 dark:text-white">
								{d.Name.replace('/dev/', '')}
							</div>
							<div class="text-sm font-normal text-gray-500 dark:text-gray-400">
								{ d.IsSystemDisk ? 'BootDisk' : '' }
							</div>
						</div>
					</TableBodyCell>
					<TableBodyCell class="p-4">{d.Size}</TableBodyCell>
					<TableBodyCell class="p-4">{d.Serial}</TableBodyCell>
					<TableBodyCell
						class="max-w-sm overflow-hidden truncate p-4 text-base font-normal text-gray-500 dark:text-gray-400 xl:max-w-xs"
						>{d.MountPoint}</TableBodyCell>
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
						<Button size="sm" class="gap-2 px-3">
							<EditOutline size="sm" /> mkfs
						</Button>
						<Button color="red" size="sm" class="gap-2 px-3">
							<TrashBinSolid size="sm" /> setMountPoint
						</Button>
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

