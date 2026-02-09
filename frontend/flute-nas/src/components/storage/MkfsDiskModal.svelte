<script lang="ts">
	import { FluteAPI } from '$lib/api';
	import type { DiskDevice } from '$lib/model';
	import { CurrentHostIP } from '$lib/vars';
	import { Button, Checkbox, Label, Modal, Spinner } from 'flowbite-svelte';
	import { createEventDispatcher } from 'svelte';
	import toast, { Toaster } from 'svelte-french-toast';

	const dispatch = createEventDispatcher();

	export let open = false;
	export let disk: DiskDevice | null = null;

	let fsType: string = 'ext4';
	let fsTypes: string[] = [];
	let confirm: boolean = false;
	let submitting: boolean = false;
	let loadingFsTypes: boolean = false;

	$: if (open) {
		fsType = 'ext4';
		fsTypes = [];
		confirm = false;
		loadSupportedFsTypes();
	}

	async function loadSupportedFsTypes() {
		loadingFsTypes = true;
		try {
			const api = new FluteAPI();
			const resp = await api.post('/v1/disk/mkfs-fstypes', {
				HostIP: $CurrentHostIP ? $CurrentHostIP : '127.0.0.1'
			});
			fsTypes = Array.isArray(resp?.data?.FsTypes) ? resp.data.FsTypes : [];
			if (fsTypes.length === 0) {
				fsType = '';
				return;
			}
			if (!fsTypes.includes(fsType)) {
				fsType = fsTypes[0];
			}
		} catch (err: any) {
			toast.error(`load supported filesystem failed: ${err?.message || 'unknown error'}`);
			fsTypes = [];
			fsType = '';
		} finally {
			loadingFsTypes = false;
		}
	}

	async function submit(e: Event) {
		e.preventDefault();
		if (!disk) return;
		if (!confirm) return;
		if (!fsType) return;

		submitting = true;
		try {
			const api = new FluteAPI();
			await api.post('/v1/disk/mkfs', {
				HostIP: $CurrentHostIP ? $CurrentHostIP : '127.0.0.1',
				Device: disk.Name,
				FsType: fsType
			});
			toast.success(`mkfs ${fsType} success: ${disk.Name}`);
			dispatch('refesh_list_msg', '');
			open = false;
		} catch (err: any) {
			toast.error(`mkfs failed: ${err?.message || 'unknown error'}`);
		} finally {
			submitting = false;
		}
	}
</script>

<Modal bind:open size="lg" autoclose={false} class="w-full">
	<form class="flex flex-col space-y-6" action="#" on:submit={submit}>
		<h3 class="mb-2 text-xl font-medium text-gray-900 dark:text-white">
			Create filesystem{disk ? ` for ${disk.Name}` : ''}
		</h3>

		<div class="space-y-2">
			<Label for="mkfs-fstype">Filesystem</Label>
			<select
				id="mkfs-fstype"
				bind:value={fsType}
				disabled={submitting || loadingFsTypes || fsTypes.length === 0}
				class="w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400"
			>
				{#if fsTypes.length === 0}
					<option value="" disabled>
						{loadingFsTypes ? 'Loading...' : 'No supported filesystems'}
					</option>
				{:else}
					{#each fsTypes as t}
						<option value={t}>{t}</option>
					{/each}
				{/if}
			</select>
		</div>

		<div class="flex items-start gap-2">
			<Checkbox bind:checked={confirm} disabled={submitting} />
			<p class="text-sm text-gray-700 dark:text-gray-300">
				I understand creating filesystem will erase existing signatures on this disk.
			</p>
		</div>

		<div class="flex justify-end gap-2">
			<Button
				color="alternative"
				type="button"
				disabled={submitting}
				on:click={() => (open = false)}
			>
				Cancel
			</Button>
			<Button type="submit" disabled={submitting || !confirm || !disk || fsTypes.length === 0}>
				{#if submitting}
					<Spinner size="4" class="mr-2" />
				{/if}
				Create
			</Button>
		</div>
	</form>
</Modal>

<Toaster />
