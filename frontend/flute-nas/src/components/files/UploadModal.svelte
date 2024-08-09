<script lang="ts">
    import { Button, Modal } from 'flowbite-svelte';
    import { Frame, ListgroupItem, Fileupload } from 'flowbite-svelte';
    import { FileProgress } from '$lib/model';
    import { createEventDispatcher } from 'svelte'; // 修改导入方式
    const dispatch = createEventDispatcher(); 

    export let open: boolean = false; // modal control


    let selectFiles :FileProgress[] = [];
    function appendSelectFile(fl :FileList) {
        
        for (let i = 0; i < fl.length; i++) {
            const file = fl[i];
            // console.log(file.name);
            selectFiles.push(new FileProgress(file, 0))
        }
        let cleanFiles :FileProgress[] = [];
        for (let i = 0; i < selectFiles.length; i++) {
            let repeated = false
            for (let j=i+1; j < selectFiles.length; j++) {
                let fi = selectFiles[i];
                let fj = selectFiles[j];
                if (fi.file.name == fj.file.name) {
                   repeated = true
                }
            }
            if (!repeated) {
                cleanFiles.push(selectFiles[i])
            }
        }
        selectFiles = cleanFiles
    }

    const dataTransfer = new DataTransfer();
    let files :FileList = dataTransfer.files;
    $: appendSelectFile(files)

    function handleDrop(e :any) {
        e.preventDefault();
        console.log(e.dataTransfer.files)
        appendSelectFile(e.dataTransfer.files)
    }

    function handleDragOver(e :any) {
        e.preventDefault();
    }

    function startUpload(e :any) {
        dispatch("selected_message", selectFiles)
    }
  </script>
  
  <Modal title="Upload Files" bind:open={open} autoclose outsideclose>
    <Fileupload id="multiple_files" multiple bind:files />
    <div
    class="flex justify-center items-center h-32 min-h-fit bg-gray-100 rounded-md"
    on:dragover={handleDragOver} 
    on:drop={handleDrop} 
    role="region" aria-label="File upload area">
        <div class="text-center">
            <p class="text-2xl font-bold">Drag the files to here</p>
            <p class="text-gray-600">release the mouse to upload files</p>
        </div>
    </div>

    <Frame tag='ul' rounded border class="divide-y divide-gray-200 dark:divide-gray-600">
        {#if (selectFiles ? selectFiles.length === 0 : true)}
            <ListgroupItem >No Selected files</ListgroupItem>
        {/if}
        {#if selectFiles}
            {#each selectFiles as item, index}
            <li
            class="py-2 px-4 w-full text-sm font-medium list-none first:rounded-t-lg last:rounded-b-lg"
            style="background-image: linear-gradient(to right, #6ee7b7 0%, #6ee7b7 {item.progress}%, #e0e7ff {item.progress+20}%, #e0e7ff 100%);"
            >{item.file.name}</li>
            {/each}
        {/if}
    </Frame>
    <svelte:fragment slot="footer">
      <Button on:click={startUpload}>Confirm</Button>
      <Button color="alternative" on:click={() => {selectFiles = []}}>Cancel</Button>
    </svelte:fragment>
  </Modal>