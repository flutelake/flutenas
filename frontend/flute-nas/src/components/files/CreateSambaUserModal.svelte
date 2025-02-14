<script lang="ts">
	import { FluteAPI } from '$lib/api';
    import { Button, Modal, Label, Input, Checkbox, P } from 'flowbite-svelte';
    import toast, { Toaster } from 'svelte-french-toast';
    import { createEventDispatcher } from 'svelte'; // 修改导入方式
    import { CurrentHostIP } from '$lib/vars';
    const dispatch = createEventDispatcher(); 
    export let open = false;
    export let dirPath :string = '/';


    let formData = {
        Username: '',
        Password: '',
        HostIP: $CurrentHostIP ? $CurrentHostIP : "127.0.0.1"
    };
    function submit(e :any) {
        e.preventDefault();

        if (formData.Username === '') {
            toast.error('Username is empty')
            return
        }
        if (formData.Password === '') {
            toast.error('Password is empty')
            return
        }
        let name = formData.Username;
        const api = new FluteAPI()
        const p = dirPath.endsWith("/") ? dirPath + name : dirPath + '/' + name
        api.post("/v1/samba-user/create", formData).then((resp :any) => {
            toast.success('Samba User ' + name + ' created')
            dispatch('refresh_samba_user_list_msg', '')
            open = false
        }).catch((err :any) => {
            console.log(err)
            toast.error('Create Samba User ' + name + ' failed')
            open = false
        })
    }
</script>
  
<Modal bind:open={open} size="xs" autoclose={false} class="w-full">
    <form class="flex flex-col space-y-6" action="#">
        <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Create Samba User</h3>
        <Label class="space-y-2">
        <span>Username</span>
        <Input type="text" name="name" placeholder="" bind:value={formData.Username} required />
        </Label>
        <Label class="space-y-2">
        <span>Password</span>
        <Input type="text" name="password" placeholder="" bind:value={formData.Password} required />
        </Label>
        <Button type="submit" class="w-full1" on:click={submit}>Confirm</Button>
    </form>
</Modal>

<Toaster />