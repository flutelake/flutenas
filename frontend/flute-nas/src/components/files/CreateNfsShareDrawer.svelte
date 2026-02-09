<script lang="ts">
	import { Drawer, Button, CloseButton, Label, Input, Popover, Select } from 'flowbite-svelte';
	import { InfoCircleSolid, TrashBinSolid } from 'flowbite-svelte-icons';
	import toast, { Toaster } from 'svelte-french-toast';
	import { sineIn } from 'svelte/easing';
	import { FluteAPI } from '$lib/api';
	import { type NFSExportAcl } from '$lib/interface';
	// import DirectoryDropdown from './DirectoryDropdown.svelte';
	import DirectoryDialog from './DirectoryDialog.svelte';
	import { onMount } from 'svelte';
	import { CurrentHostIP } from '$lib/vars';
	import { createEventDispatcher } from 'svelte';
	const dispatch = createEventDispatcher();

	export let selectedPath = '';
	export let hidden = true;
	export let selectDirModalFlag = false;
	// $:  = !open;
	let transitionParamsRight = {
		x: 320,
		duration: 200,
		easing: sineIn
	};
	let formData: {
		Name: string;
		Path: string;
		Pseudo: string;
		DefaultACL: string;
		Acls: NFSExportAcl[];
	} = {
		Name: '',
		Path: '',
		Pseudo: '',
		DefaultACL: 'None',
		Acls: []
	};

	let acls: NFSExportAcl[] = [
		{
			IPRange: '',
			Permission: ''
		}
	];

	let permissions = [
		{ value: 'RW', name: 'Read & Write' },
		{ value: 'RO', name: 'Read Only' },
		{ value: 'None', name: 'Deny Access' }
	];

	function onClickAddAclInput() {
		acls.push({
			IPRange: '',
			Permission: ''
		});
		acls = acls;
	}

	function onClickDelAclInput(idx: number) {
		if (acls.length <= 1) {
			// 不允许全部删除掉
			return;
		}
		if (idx >= 0 && idx < acls.length) {
			acls.splice(idx, 1); // 从 idx 开始删除 1 个元素
		}
		acls = acls;
	}

	function submit() {
		formData.Acls = acls;
		formData.Path = selectedPath; // 确保使用选择的路径

		// 检查表单参数
		if (!formData.Name.trim()) {
			toast.error('The share name cannot be empty.');
			return;
		}

		if (!formData.Path.trim()) {
			toast.error('The path cannot be empty.');
			return;
		}

		if (!formData.Pseudo.trim()) {
			toast.error('Pseudo path cannot be empty.');
			return;
		}

		// 检查ACL配置
		for (let i = 0; i < formData.Acls.length; i++) {
			const acl = formData.Acls[i];
			if (!acl.IPRange.trim()) {
				toast.error(`The IP range of the ${i + 1}th NFS client cannot be empty.`);
				return;
			}
			if (!acl.Permission) {
				toast.error(`The permission of the ${i + 1}th NFS client cannot be empty.`);
				return;
			}
		}

		const api = new FluteAPI();
		api
			.post('/v1/nfs-share/create', {
				...formData,
				HostIP: $CurrentHostIP ? $CurrentHostIP : '127.0.0.1'
			})
			.then((resp) => {
				// 提交成功后的处理
				hidden = true; // 关闭抽屉
				toast.success('NFS share created successfully');
				// 触发刷新列表事件
				dispatch('refresh_nfs_share_list_msg');
			})
			.catch((err) => {
				toast.error('Failed to create NFS share: ' + err.message);
			});
	}
</script>

<Drawer
	placement="right"
	transitionType="fly"
	{transitionParamsRight}
	bind:hidden
	id="sidebar3"
	activateClickOutside={false}
	width="w-2/5"
>
	<div class="flex items-center">
		<h5
			id="drawer-label"
			class="mb-6 inline-flex items-center text-base font-semibold uppercase text-gray-500 dark:text-gray-400"
		>
			<InfoCircleSolid class="me-2.5 h-5 w-5" />Create NFS Share
		</h5>
		<CloseButton on:click={() => (hidden = true)} class="mb-4 dark:text-white" />
	</div>
	<form action="#" class="mb-6">
		<div class="mb-6">
			<Label for="ShareName" class="mb-2 block">Share Name</Label>
			<Input id="ShareName" name="ShareName" bind:value={formData.Name} required placeholder="" />
		</div>
		<div class="mb-6">
			<Label for="Pseudo" class="mb-2 block">Pseudo Path</Label>
			<Input id="Pseudo" name="Pseudo" bind:value={formData.Pseudo} required placeholder="" />
		</div>
		<div class="mb-6">
			<Label for="Path" class="mb-2">Shared Dir Path</Label>
			<button on:click={() => (selectDirModalFlag = true)}
				><Input
					id="Path"
					name="Path"
					value={selectedPath}
					required
					placeholder=""
					readonly
				/></button
			>
		</div>
		<div class="mb-6">
			<Label for="DefaultACL" class="mb-2">Default ACL</Label>
			<Select
				id="DefaultACL"
				items={permissions}
				bind:value={formData.DefaultACL}
				placeholder="Choose Default Permission"
			/>
		</div>
		<!-- configure NFS clients with ACLs -->
		{#if formData.DefaultACL != 'RW'}
			<div class="mb-6">
				<Label for="client" class="mb-2">Client Access Control</Label>
				{#each acls as acl, index}
					<div class="grid grid-cols-5 content-start gap-6">
						<div class="col-span-2">
							<Input
								id="IPRange"
								name="IPRange"
								bind:value={acl.IPRange}
								placeholder="Client IPs"
							/>
						</div>
						<div class="col-span-2">
							<Select
								class="mb-2"
								items={permissions}
								bind:value={acl.Permission}
								placeholder="Choose Permission"
							/>
						</div>
						<div class="col-span-1">
							{#if index == 0}
								<Button
									color="red"
									class="w-full"
									on:click={() => onClickDelAclInput(index)}
									disabled
								>
									<TrashBinSolid size="sm" /> Del
								</Button>
							{:else}
								<Button color="red" class="w-full" on:click={() => onClickDelAclInput(index)}>
									<TrashBinSolid size="sm" /> Del
								</Button>
							{/if}
						</div>
					</div>
				{/each}
				<Button on:click={() => onClickAddAclInput()} class="w-full">+ Add Client</Button>
			</div>
		{/if}

		<Button on:click={() => submit()} class="w-full">Submit</Button>
	</form>
</Drawer>

<DirectoryDialog
	bind:open={selectDirModalFlag}
	bind:selectedPath
	on:refresh_nfs_share_list_msg={() => console.log('1112')}
></DirectoryDialog>
