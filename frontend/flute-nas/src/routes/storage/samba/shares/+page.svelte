<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Breadcrumb,
		BreadcrumbItem,
		Button,
		Checkbox,
		Drawer,
		Heading,
		Input,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell,
		Toolbar,
		ToolbarButton,
		Badge,
		Hr,
		GradientButton
	} from 'flowbite-svelte';
	import {
		CogSolid,
		DotsVerticalOutline,
		EditOutline,
		ExclamationCircleSolid,
		TrashBinSolid,
		RefreshOutline
	} from 'flowbite-svelte-icons';
	import type { ComponentType } from 'svelte';
	import { sineIn } from 'svelte/easing';
	import MetaTag from '../../../../components/MetaTag.svelte';
	import { FluteAPI } from '$lib/api';
	import { DiskDevice } from '$lib/model';
	// import SetMountPointModal from '../../../../components/storage/SetMountPointModal.svelte';
	export let node: string = 'localhost';
	import { CurrentHostIP } from '$lib/vars';
	import { type SambaShare, type UserPermission } from '$lib/interface';
	import { formatDateTime } from '$lib/index';
	// import CreateSambaShareModal from '../../../../components/files/CreateSambaShareModal.svelte';
	import UpdateSambaUserModal from '../../../../components/files/UpdateSambaUserModal.svelte';
	import DeleteSambaUserModal from '../../../../components/files/DeleteSambaUserModal.svelte';
	import CreateSambaShareDrawer from '../../../../components/files/CreateSambaShareDrawer.svelte';
	// import DirectoryDialog from '../../../../components/files/DirectoryDialog.svelte';

	// let hidden: boolean = true; // modal control
	let loading: boolean = false;

	// const toggle = (component: ComponentType) => {
	// 	hidden = !hidden;
	// };

	const path: string = '/samba/share';
	const description: string = 'Samba Share - FluteNAS Web Console';
	const title: string = 'FluteNAS Web Console - Samba Share';
	const subtitle: string = 'Samba Share';
	// let transitionParams = {
	// 	x: 320,
	// 	duration: 200,
	// 	easing: sineIn
	// };

	let shares: SambaShare[] = [];
	$: refreshList($CurrentHostIP);
	function refreshList(ip: string = $CurrentHostIP) {
		// console.log('refreshList, ip: ', ip)
		if (loading) {
			// 防重复点击
			return;
		}
		loading = true;
		console.log(ip);
		const api = new FluteAPI();
		api
			.post('/v1/samba-share/list', {})
			.then((resp) => {
				shares = resp.data.Shares;
				loading = false;
			})
			.catch((err) => {
				console.log(err);
				loading = false;
			});
	}

	let createShareModalFlag: boolean = true;
	function onClickCreateShare() {
		console.log('open Create Samba Share Dialog');
		createShareModalFlag = false;
	}

	let smbStatus: boolean = false;
	function checkSmbStatus() {
		const api = new FluteAPI();
		api
			.post('/v1/samba-share/status', { HostIP: $CurrentHostIP ? $CurrentHostIP : '127.0.0.1' })
			.then((resp) => {
				smbStatus = resp.data.Actived;
			})
			.catch((err) => {});
	}

	let deleteUserModalFlag: boolean = false;
	function onClickDeleteShare(idx: number) {
		// console.log('open Update Samba User Dialog')
		// deleteUserModalFlag = true
		// selectUser = users[idx]
	}

	function formatUserPermission(us: UserPermission[]) {
		let ustr = '';
		us.forEach((u) => {
			ustr += u.Username + ' : ' + formatPermission(u.Permission);
		});
		return ustr;
	}

	function formatPermission(str: string) {
		switch (str) {
			case 'r':
				return 'Read Only';
			case 'rw':
				return 'Read & Write';
			default:
				return '';
		}
	}
	let selectedPath = '';

	onMount(() => {
		checkSmbStatus();
	});
</script>

<MetaTag {path} {description} {title} {subtitle} />

<main class="relative h-full w-full overflow-y-auto bg-white dark:bg-gray-800">
	<div class="p-4">
		<Breadcrumb class="mb-5">
			<BreadcrumbItem href="/" home>Home</BreadcrumbItem>
			<BreadcrumbItem>Samba</BreadcrumbItem>
			<BreadcrumbItem>Shares</BreadcrumbItem>
		</Breadcrumb>
		<Heading tag="h1" class="text-xl font-semibold text-gray-900 sm:text-2xl dark:text-white">
			Samba Shares
		</Heading>
		<!-- <p class="mb-6 text-lg font-normal text-gray-500 dark:text-gray-400 sm:text-xl">
			Active
			</p> -->

		<Toolbar embedded class="w-full py-4 text-gray-500 dark:text-gray-400">
			<div class="items-center justify-between gap-3 space-y-4 sm:flex sm:space-y-0">
				<div class="flex items-center space-x-4">
					<!-- <GradientButton color="purpleToBlue" ></GradientButton> -->
					<GradientButton color="pinkToOrange" on:click={onClickCreateShare}>Create</GradientButton>
				</div>
				<div class="flex items-center space-x-4">
					<!-- <GradientButton color="purpleToBlue" ></GradientButton> -->
					<!-- <GradientButton color="pinkToOrange" >smb status: active</GradientButton> -->
					{#if smbStatus}
						<p>Samba Status is Actived</p>
					{:else}
						<p>Samba Status not Active</p>
					{/if}
				</div>
			</div>
			<div slot="end" class="space-x-2">
				<!-- on:click={() => toggle("")} -->
				<Button
					pill={true}
					class="bg-gradient-to-br from-purple-600 to-blue-500 !p-2 text-white shadow-blue-500/50 hover:bg-gradient-to-bl focus:ring-blue-300 dark:shadow-blue-800/80 dark:focus:ring-blue-800"
					on:click={() => {
						refreshList($CurrentHostIP);
					}}
				>
					{#if loading}
						<RefreshOutline class="spin-fast h-6 w-6" />&nbsp; Loading... Please wait
					{:else}
						<RefreshOutline class="h-6 w-6" />
					{/if}
				</Button>
				<!-- <Button class="whitespace-nowrap" >Add new product</Button> -->
			</div>
		</Toolbar>
	</div>
	<Table>
		<TableHead class="border-y border-gray-200 bg-gray-100 dark:border-gray-700">
			<TableHeadCell class="w-4 p-4"><Checkbox /></TableHeadCell>
			{#each ['Index', 'ShareName', 'Shared Dir Path', 'Pseudo Path', 'Status', 'Users', 'CreateAt', 'Action'] as title}
				<TableHeadCell class="ps-4 font-normal">{title}</TableHeadCell>
			{/each}
		</TableHead>
		<TableBody>
			{#each shares as s, index}
				<TableBodyRow class="text-base">
					<TableBodyCell class="w-4 p-4"><Checkbox /></TableBodyCell>
					<TableBodyCell class="p-4">{s.ID}</TableBodyCell>
					<TableBodyCell class="p-4">{s.Name}</TableBodyCell>
					<TableBodyCell class="p-4">{s.Path}</TableBodyCell>
					<TableBodyCell class="p-4">{s.Pseudo}</TableBodyCell>
					<TableBodyCell class="p-4">{s.Status}</TableBodyCell>
					<TableBodyCell class="p-4">
						<div class="space-y-1">
							{#each s.Users as u}
								<div
									class="flex items-center justify-between rounded-md border border-blue-200 bg-blue-50 px-2 py-1 text-xs dark:border-blue-800 dark:bg-blue-900/20"
								>
									<span class="font-mono font-medium text-blue-800 dark:text-blue-300"
										>{u.Username}</span
									>
									<span class="ml-2 font-semibold text-blue-600 dark:text-blue-400"
										>{formatPermission(u.Permission)}</span
									>
								</div>
							{/each}
						</div>
					</TableBodyCell>
					<TableBodyCell class="p-4">{formatDateTime(s.CreatedAt)}</TableBodyCell>
					<TableBodyCell class="p-4">
						<!-- <Button size="sm" class="gap-2 px-3" on:click={() => onClickUpdateUser(index)}>
							<EditOutline size="sm" /> Edit PWD
						</Button> -->
						<Button
							color="red"
							size="sm"
							class="gap-2 px-3"
							on:click={() => onClickDeleteShare(index)}
						>
							<TrashBinSolid size="sm" /> Delete
						</Button>
					</TableBodyCell>
					<!-- <TableBodyCell class="p-4">{}</TableBodyCell> -->
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>
	<Hr />
	<!-- <Hr /> -->
	<div class="p-4">
		<Heading
			tag="h1"
			size="xl"
			class="mb-3 text-3xl font-bold text-gray-900 sm:text-4xl sm:leading-none sm:tracking-tight dark:text-white"
		>
			FAQ
		</Heading>
		<p class="mb-6 text-lg font-normal text-gray-500 sm:text-xl dark:text-gray-400">
			frequently asked questions
		</p>

		<!-- <div class="lg:columns-1 gap-8 space-y-10">
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
		</div> -->
	</div>
</main>

<!-- <SetMountPointModal bind:open={openSetMountPointModal} bind:node={node} bind:disk={currentDevice} on:refesh_list_msg={()=>refreshList($CurrentHostIP)}></SetMountPointModal> -->

<!-- <UpdateSambaUserModal bind:open={updateUserModalFlag} bind:user={selectUser} on:refresh_samba_user_list_msg={()=>refreshList()}></UpdateSambaUserModal>
<DeleteSambaUserModal bind:open={deleteUserModalFlag} bind:user={selectUser} on:refresh_samba_user_list_msg={()=>refreshList()}></DeleteSambaUserModal> -->

<!-- <CreateSambaShareModal bind:open={createShareModalFlag} on:refresh_samba_user_list_msg={()=>refreshList()}></CreateSambaShareModal> -->
<CreateSambaShareDrawer
	bind:hidden={createShareModalFlag}
	bind:selectedPath
	on:refresh_samba_user_list_msg={() => refreshList()}
></CreateSambaShareDrawer>
