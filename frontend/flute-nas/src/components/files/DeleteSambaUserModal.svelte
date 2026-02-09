<script lang="ts">
	import { FluteAPI } from '$lib/api';
	import { Button, Modal } from 'flowbite-svelte';
	import { ExclamationCircleOutline } from 'flowbite-svelte-icons';
	import toast, { Toaster } from 'svelte-french-toast';
	import { createEventDispatcher } from 'svelte'; // 修改导入方式
	import type { SambaUser } from '$lib/interface';
	const dispatch = createEventDispatcher();
	export let open = false;
	export let user: SambaUser = {
		ID: 0,
		Username: '',
		Password: '',
		Status: '',
		CreatedAt: new Date(),
		Permission: ''
	};

	function submit(e: any) {
		e.preventDefault();
		if (!user.ID) {
			open = false;
			return;
		}
		const api = new FluteAPI();
		api
			.post('/v1/samba-user/delete', { ID: user.ID })
			.then((resp) => {
				toast.success('delete success');
				dispatch('refresh_samba_user_list_msg', '');
				open = false;
			})
			.catch((e) => {
				toast.error('delete failed');
				open = false;
				console.log(e);
			});
	}
</script>

<Modal bind:open size="xs" autoclose>
	<div class="text-center">
		<ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-200" />
		<h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
			Are you sure you want to delete this samba user: {user.Username} ?
		</h3>
		<Button color="red" class="me-2" on:click={submit}>Yes, I'm sure</Button>
		<Button color="alternative">No, cancel</Button>
	</div>
</Modal>

<Toaster />
