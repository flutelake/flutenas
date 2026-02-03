<script lang="ts">
	import { onMount } from 'svelte';
	import { Breadcrumb, BreadcrumbItem, Button, Checkbox, Heading, Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Toolbar, Hr, GradientButton } from 'flowbite-svelte';
	import { TrashBinSolid, RefreshOutline, LinkOutline, ClipboardSolid, CheckOutline } from 'flowbite-svelte-icons';
	import MetaTag from '../../../../components/MetaTag.svelte';
	import { FluteAPI } from '$lib/api';
	import { CurrentHostIP } from '$lib/vars';
	import { type NFSExport, type NFSExportAcl } from '$lib/interface'
	import { formatDateTime } from '$lib/index';
	import CreateNfsShareDrawer from '../../../../components/files/CreateNfsShareDrawer.svelte';
import DeleteNfsShareModal from '../../../../components/files/DeleteNfsShareModal.svelte';
import MountCommandModal from '../../../../components/files/MountCommandModal.svelte';

	// let hidden: boolean = true; // modal control
	let loading: boolean = false;

	const path: string = '/nfs/share';
  	const description: string = 'NFS Share - FluteNAS Web Console';
	const title: string = 'FluteNAS Web Console - NFS Share';
	const subtitle: string = 'NFS Share';
	// let transitionParams = {
	// 	x: 320,
	// 	duration: 200,
	// 	easing: sineIn
	// };

	let shares :NFSExport[] = [];
	function refreshList(ip :string = "127.0.0.1") {
		// console.log('refreshList, ip: ', ip)
		if (loading) {
			// 防重复点击
			return
		}
		loading = true;
		console.log(ip)
		const api = new FluteAPI();
        api.post("/v1/nfs-share/list", {"HostIP": ip}).then(resp => {
			console.log('API response:', resp); // 添加调试信息
			shares = Array.isArray(resp.data?.Exports) ? resp.data.Exports : [];
			console.log('Shares after update:', shares); // 添加调试信息
			loading = false;
        }).catch(err => {
            console.log(err)
			shares = []; // Ensure shares is always an array
			loading = false;
        })
	}

	let createShareModalFlag :boolean = true
	function onClickCreateShare() {
		console.log('open Create NFS Share Dialog')
		createShareModalFlag = false
	}

	let nfsStatus :boolean = false
	function checkNfsStatus() {
		const api = new FluteAPI();
		api.post("/v1/nfs-share/status", {"HostIP": $CurrentHostIP ? $CurrentHostIP : "127.0.0.1"}).then(resp => {
			nfsStatus = resp.data.Actived;
        }).catch(err => {
        })
	}

	let deleteUserModalFlag :boolean = false
let selectedShare :NFSExport | null = null
let mountCommandModalFlag :boolean = false
let selectedMountShare :NFSExport | null = null
	
function onClickDeleteShare(idx :number) {
	selectedShare = shares[idx];
	deleteUserModalFlag = true;
}

function onClickMountCommand(idx :number) {
	selectedMountShare = shares[idx];
	mountCommandModalFlag = true;
}

	function formatUserPermission(us :NFSExportAcl[]) {
		if (!Array.isArray(us) || us.length === 0) {
			return '';
		}
		
		// If only one ACL, return simple format
		if (us.length === 1) {
			return `${us[0].IPRange} : ${formatPermission(us[0].Permission)}`;
		}
		
		// For multiple ACLs, create a formatted list with better separation
		const aclEntries = us.map((u) => {
			return `• ${u.IPRange} : ${formatPermission(u.Permission)}`;
		});
		
		// Join with line breaks for better readability, add spacing between entries
		return aclEntries.join(' \\n ');
	}

	function formatPermission(str :string) {
		switch (str) {
			case 'RO':
				return 'Read Only'
			case 'RW':
				return 'Read & Write'
			case 'None':
				return 'Deny Access'
			default:
				return ''
		}
	}
	let selectedPath = ""

	onMount(() => {
		checkNfsStatus()
		refreshList($CurrentHostIP)
	})
</script>

<MetaTag {path} {description} {title} {subtitle} />

<main class="relative h-full w-full overflow-y-auto bg-white dark:bg-gray-800">
	<div class="p-4">
		<Breadcrumb class="mb-5">
			<BreadcrumbItem href="/" home>Home</BreadcrumbItem>
			<BreadcrumbItem>NFS</BreadcrumbItem>
			<BreadcrumbItem>Shares</BreadcrumbItem>
		</Breadcrumb>
		<Heading tag="h1" class="text-xl font-semibold text-gray-900 dark:text-white sm:text-2xl">
			NFS Shares
		</Heading>
		<!-- <p class="mb-6 text-lg font-normal text-gray-500 dark:text-gray-400 sm:text-xl">
			Active
			</p> -->
		
		<Toolbar embedded class="w-full py-4 text-gray-500 dark:text-gray-400">
			<div class="items-center justify-between gap-3 space-y-4 sm:flex sm:space-y-0">
				<div class="flex items-center space-x-4">
					<!-- <GradientButton color="purpleToBlue" ></GradientButton> -->
					<GradientButton color="pinkToOrange" on:click={onClickCreateShare} >Create</GradientButton>
				</div>
				<div class="flex items-center space-x-4">
					<!-- <GradientButton color="purpleToBlue" ></GradientButton> -->
					<!-- <GradientButton color="pinkToOrange" >nfs status: active</GradientButton> -->
					{#if nfsStatus }
					<p>NFS Status is Active</p>
					{:else}
					<p>NFS Status not Active</p>
					{/if}
				</div>
			</div>
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
			{#each ['ID', 'ShareName', 'Shared Dir Path', 'Pseudo Path', 'Status', 'Default ACL', 'Access Control', 'CreateAt', 'Mount', 'Action'] as title}
				<TableHeadCell class="ps-4 font-normal" style={title === 'Access Control' ? 'min-width: 200px' : ''}>{title}</TableHeadCell>
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
						<!-- <span class="text-xs font-medium px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded-md"> -->
						<div class="flex items-center justify-between bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md px-2 py-1 text-xs">
							<span class="font-mono text-blue-800 dark:text-blue-300 font-medium">{formatPermission(s.DefaultACL)}</span>
						</div>
						<!-- </span> -->
					</TableBodyCell>
					<TableBodyCell class="p-4">
						<div class="space-y-1">
							{#if s.Acls && s.Acls.length > 0}
								{#each s.Acls as acl}
									<div class="flex items-center justify-between bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md px-2 py-1 text-xs">
										<span class="font-mono text-blue-800 dark:text-blue-300 font-medium">{acl.IPRange}</span>
										<span class="text-blue-600 dark:text-blue-400 ml-2 font-semibold">{formatPermission(acl.Permission)}</span>
									</div>
								{/each}
							{:else}
								<span class="text-gray-400 text-xs italic">No ACLs configured</span>
							{/if}
						</div>
					</TableBodyCell>
					<TableBodyCell class="p-4">{formatDateTime(s.CreatedAt)}</TableBodyCell>
					<TableBodyCell class="p-4">
						<Button color="blue" size="sm" class="gap-2 px-3" on:click={() => onClickMountCommand(index)}>
							<LinkOutline size="sm" /> Mount
						</Button>
					</TableBodyCell>
					<TableBodyCell class="p-4">
						<!-- <Button size="sm" class="gap-2 px-3" on:click={() => onClickUpdateUser(index)}>
							<EditOutline size="sm" /> Edit PWD
						</Button> -->
						<Button color="red" size="sm" class="gap-2 px-3" on:click={() => onClickDeleteShare(index)}>
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
					Default ACL?
				</h3>
				<p class="text-gray-600 dark:text-gray-400">
					The default access control permissions apply to all users who access the shared directory. Therefore, in most cases, you should set the default access policy to “Deny access” and then configure access permissions separately for each client IP.
				</p>
				<p class="text-gray-600 dark:text-gray-400">
					If you set the Default ACL to “Read & Write”, it means everyone has permission to access this shared directory, so there is no need to configure access control for specific clients.
				</p>
			</div>

			<div class="space-y-4">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">
					Client Access Control - Client IP formats
				</h3>
				<p class="text-gray-600 dark:text-gray-400">
					Client IPs support the following formats:
				</p>
				<ul class="list-disc list-inside space-y-1 text-gray-600 dark:text-gray-400">
					<li>
						Single IP:
						<span class="font-mono">192.168.1.10</span>
					</li>
					<li>
						IP list:
						<span class="font-mono">192.168.1.10, 192.168.1.11</span>
						<span class="text-xs text-gray-500 dark:text-gray-500">(separate with commas)</span>
					</li>
					<li>
						Wildcard:
						<span class="font-mono">*</span>
						matches any client, or
						<span class="font-mono">192.168.1.*</span>
						for prefix match
					</li>
					<li>
						CIDR:
						<span class="font-mono">192.168.1.0/24</span>
					</li>
				</ul>
			</div>
		</div>
	</div>
</main>

<!-- <SetMountPointModal bind:open={openSetMountPointModal} bind:node={node} bind:disk={currentDevice} on:refesh_list_msg={()=>refreshList($CurrentHostIP)}></SetMountPointModal> -->

<!-- <UpdateNfsShareModal bind:open={updateUserModalFlag} bind:user={selectUser} on:refresh_nfs_share_list_msg={()=>refreshList()}></UpdateNfsShareModal>
<DeleteNfsShareModal bind:open={deleteUserModalFlag} bind:user={selectUser} on:refresh_nfs_share_list_msg={()=>refreshList()}></DeleteNfsShareModal> -->

<!-- <CreateNfsShareModal bind:open={createShareModalFlag} on:refresh_nfs_share_list_msg={()=>refreshList()}></CreateNfsShareModal> -->
<CreateNfsShareDrawer bind:hidden={createShareModalFlag} bind:selectedPath={selectedPath} on:refresh_nfs_share_list_msg={()=>refreshList()}></CreateNfsShareDrawer>

<!-- Delete Confirmation Modal -->
<DeleteNfsShareModal 
  bind:open={deleteUserModalFlag}
  shareName={selectedShare ? selectedShare.Name : ''} 
  shareId={selectedShare ? selectedShare.ID : 0}
  on:refresh_nfs_share_list_msg={() => {
    refreshList($CurrentHostIP);
  }}
/>

<!-- Mount Command Modal -->
<MountCommandModal 
  bind:open={mountCommandModalFlag}
  share={selectedMountShare}
/>
