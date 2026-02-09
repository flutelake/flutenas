<script lang="ts">
	import { FluteAPI } from '$lib/api';
	import { Button, Modal, Label, Input, Checkbox, P } from 'flowbite-svelte';
	import toast, { Toaster } from 'svelte-french-toast';
	import { createEventDispatcher } from 'svelte'; // 修改导入方式
	import { type SambaUser } from '$lib/interface';
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

		if (user.Username === '') {
			toast.error('Username is empty');
			return;
		}
		if (user.Password === '') {
			toast.error('Password is empty');
			return;
		}
		let name = user.Username;
		const api = new FluteAPI();
		api
			.post('/v1/samba-user/update', user)
			.then((resp: any) => {
				toast.success('Samba User ' + name + ' updated');
				dispatch('refresh_samba_user_list_msg', '');
				open = false;
			})
			.catch((err: any) => {
				console.log(err);
				toast.error('Update Samba User ' + name + ' failed');
				open = false;
			});
	}
</script>

<Modal bind:open size="xs" autoclose={false} class="w-full">
	<form class="flex flex-col space-y-6" action="#">
		<h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">
			Update Samba User Password
		</h3>
		<Label class="space-y-2">
			<span>Username</span>
			<Input disabled type="text" name="name" placeholder="" bind:value={user.Username} required />
		</Label>
		<Label class="space-y-2">
			<span>Password</span>
			<Input type="text" name="password" placeholder="" bind:value={user.Password} required />
		</Label>
		<Button type="submit" class="w-full1" on:click={submit}>Confirm</Button>
	</form>
</Modal>

<Toaster />
