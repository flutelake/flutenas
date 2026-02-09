<script lang="ts">
	import { FluteAPI } from '$lib/api';
	import { Button, Modal } from 'flowbite-svelte';
	import { ExclamationCircleOutline } from 'flowbite-svelte-icons';
	import toast, { Toaster } from 'svelte-french-toast';
	import { createEventDispatcher } from 'svelte';
	import { CurrentHostIP } from '$lib/vars';

	const dispatch = createEventDispatcher();
	export let open = false;
	export let shareName = '';
	export let shareId = 0;

	function submit(e: Event) {
		e.preventDefault();
		if (shareId === 0) {
			open = false;
			return;
		}

		const api = new FluteAPI();
		api
			.post('/v1/nfs-share/delete', {
				ID: shareId,
				HostIP: $CurrentHostIP ? $CurrentHostIP : '127.0.0.1'
			})
			.then((resp) => {
				toast.success('NFS share deleted successfully');
				dispatch('refresh_nfs_share_list_msg', '');
				open = false;
			})
			.catch((e: any) => {
				toast.error('Failed to delete NFS share');
				open = false;
				console.log(e);
			});
	}
</script>

<Modal bind:open size="xs" autoclose>
	<div class="text-center">
		<ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-200" />
		<h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
			Are you sure you want to delete this NFS share: {shareName} ?
		</h3>
		<Button color="red" class="me-2" on:click={submit}>Yes, I'm sure</Button>
		<Button color="alternative">No, cancel</Button>
	</div>
</Modal>

<Toaster />
