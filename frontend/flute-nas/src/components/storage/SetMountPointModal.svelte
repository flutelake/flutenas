<script lang="ts">
	import { FluteAPI } from '$lib/api';
    import { Button, Modal, Label, Input, Checkbox } from 'flowbite-svelte';
    import toast, { Toaster } from 'svelte-french-toast';
    import { createEventDispatcher } from 'svelte'; // 修改导入方式
	import type { DiskDevice } from '$lib/model';
    import { CurrentHostIP } from '$lib/vars';
    const dispatch = createEventDispatcher(); 
    export let open = false;
    export let node :string = 'localhost';
    export let disk :DiskDevice;

    function submit(e :any) {
        e.preventDefault();

        // if (disk.SpecMountPoint === '') {
        //     // toast.error('mount-point is empty')
        //     return
        // }
        let mpoint = disk.SpecMountPoint;
        const api = new FluteAPI()
        
        api.post("/v1/disk/set-mountpoint", {'HostIP': $CurrentHostIP ? $CurrentHostIP : "127.0.0.1",'Node': node, 'Device': disk.Name, 'UUID': disk.UUID, 'Path': mpoint}).then((resp :any) => {
            toast.success('dir ' + mpoint + ' created')
            dispatch('refesh_list_msg', '')
            open = false
        }).catch((err :any) => {
            console.log(err)
            toast.error('set mount-point: ' + mpoint + ' failed')
            open = false
        })
    }
</script>
  
<Modal bind:open={open} size="xs" autoclose={false} class="w-full">
    <form class="flex flex-col space-y-6" action="#">
        <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">Set MountPoint for {disk.Name}</h3>
        <Label class="space-y-2">
        <span>MountPoint</span>
        <Input type="text" name="mount-point" placeholder="eg: /abc" bind:value={disk.SpecMountPoint} required />
        </Label>
        <Button type="submit" class="w-full1" on:click={submit}>Confirm</Button>
    </form>
</Modal>

<Toaster />