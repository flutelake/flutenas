<script lang="ts">
	import { FluteAPI } from '$lib/api';
	import { Button, Modal, Label, Input, Checkbox } from 'flowbite-svelte';
	import toast, { Toaster } from 'svelte-french-toast';
	import { createEventDispatcher } from 'svelte'; // 修改导入方式
	const dispatch = createEventDispatcher();
	export let open = false;
	export let dirPath: string = '/';

	let formData = {
		name: ''
	};
	function submit(e: any) {
		e.preventDefault();

		if (formData.name === '') {
			toast.error('dir name is empty');
			return;
		}
		let name = formData.name;
		const api = new FluteAPI();
		const p = dirPath.endsWith('/') ? dirPath + name : dirPath + '/' + name;
		api
			.post('/v1/files/createdir', { Path: p })
			.then((resp: any) => {
				toast.success('dir ' + name + ' created');
				dispatch('refesh_list_msg', '');
				open = false;
			})
			.catch((err: any) => {
				console.log(err);
				toast.error('create dir ' + name + ' failed');
				open = false;
			});
	}
</script>

<Modal bind:open size="xs" autoclose={false} class="w-full">
	<form class="flex flex-col space-y-6" action="#">
		<h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Create Directory</h3>
		<Label class="space-y-2">
			<span>Name</span>
			<Input type="text" name="name" placeholder="" bind:value={formData.name} required />
		</Label>
		<Button type="submit" class="w-full1" on:click={submit}>Confirm</Button>
	</form>
</Modal>

<Toaster />
