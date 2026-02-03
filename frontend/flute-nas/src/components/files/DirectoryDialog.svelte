<script lang="ts">
    import { Modal, Button } from 'flowbite-svelte';
    import Tree from './Tree.svelte';
    import { type DirTreeNode } from '$lib/interface';
    
    // export let onSelect: (path: string) => void;
    
    export let open = false;
    export let selectedPath = '';
    const sampleData: DirTreeNode[] = [{
		name: '/',
		path: '/',
		children: []
	}]


    function handleConfirm() {
        if (selectedPath.startsWith('//')) {
            selectedPath = selectedPath.replace('//','/')
        }
        open = false;
    }

    function handleCancel() {
        selectedPath = ''
        open = false;
    }
</script>

<Modal title="Select Directory" bind:open={open}>
    <div class="p-4">
        <Tree data={sampleData} bind:selectedPath />
    </div>
    
    <div slot="footer" class="flex justify-end gap-2">
        <Button color="alternative" on:click={handleCancel}>取消</Button>
        <Button color="primary" on:click={handleConfirm}>确定</Button>
    </div>
</Modal>
