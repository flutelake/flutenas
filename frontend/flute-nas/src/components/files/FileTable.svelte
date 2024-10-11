<script lang="ts">
	import { FileEntry } from '$lib/model';
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
  import { A } from 'flowbite-svelte';
  import { FolderOpenSolid, RefreshOutline } from 'flowbite-svelte-icons';
  import { formatSize } from '$lib/index';
  import DeleteModal from './DeleteModal.svelte';
  import { createEventDispatcher } from 'svelte';
	import { FluteAPI } from '$lib/api';
    const dispatch = createEventDispatcher(); 

  export let files: FileEntry[] = [];
  export let dirPath : string = '/';
  export let loading : boolean = false;

  function forward() {
    const pathParts = dirPath.split('/').filter(Boolean);
    pathParts.pop(); // 移除最后一个部分
    dirPath = '/' + pathParts.join('/'); 
  }

  let downlodFilename :string = ''
  function handleDownload() {
		let p = dirPath.endsWith('/') ? dirPath + downlodFilename : dirPath + '/' + downlodFilename;
    const api = new FluteAPI();
    api.post("/v1/files/download", {'Path': p}).then((resp :any) => {
      console.log(resp.data) 
      // resp.data.Location
      if (resp.data.Location) {
        // 重定向到 Location URL
        window.location.href = resp.data.Location;
      } else {
        console.error('No Location field in response data');
      }
    }).then(err => {
      console.log(err);
    })
	}

  const dateTimeOptions = { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' } as const;
  let openDeleteModal :boolean = false;
  let deleteFileName :string = '';
</script>

<Table hoverable={true}>
    <TableHead>
      <TableHeadCell>FileName</TableHeadCell>
      <TableHeadCell>Size</TableHeadCell>
      <TableHeadCell>LastModify</TableHeadCell>
      <TableHeadCell>Kind</TableHeadCell>
      <TableHeadCell>Actions</TableHeadCell>
    </TableHead>
    <TableBody tableBodyClass="divide-y">
        {#if loading}
        <TableBodyRow>
          <TableBodyCell><A href="#" on:click={forward}><RefreshOutline class="spin-fast"></RefreshOutline>&nbsp;Loading... Please wait </A></TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
        </TableBodyRow>
        {/if}
        {#if dirPath != '' && dirPath != '/'}
        <TableBodyRow>
          <TableBodyCell><A href="#" on:click={forward}><FolderOpenSolid></FolderOpenSolid>&nbsp;...</A></TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
          <TableBodyCell>&nbsp;</TableBodyCell>
        </TableBodyRow>
        {/if}
        {#each files as file}
        <TableBodyRow>
            {#if file.isDir }
            <TableBodyCell><A href="#" on:click={() => {dirPath = dirPath.endsWith('/')? dirPath + file.name : dirPath + '/' + file.name}}><FolderOpenSolid></FolderOpenSolid>&nbsp;{file.name}</A></TableBodyCell>
            {:else}
            <TableBodyCell>&nbsp;{file.name}</TableBodyCell>
            {/if}
            <TableBodyCell>{file.isDir ? '' : formatSize(file.size)}</TableBodyCell>
            <TableBodyCell>{new Date(file.lastMod).toLocaleString('zh-CN', dateTimeOptions)}</TableBodyCell>
            <TableBodyCell>{file.isDir ? '' : file.kind}</TableBodyCell>
            <TableBodyCell>
            <A href="#" on:click={() => {downlodFilename = file.name; handleDownload()}}>Download</A>
            <A href="#" on:click={() => {deleteFileName = file.name; openDeleteModal = true;}}>Delete</A>
            <!-- <a href="/tables" class="font-medium text-primary-600 hover:underline dark:text-primary-500">Review</a> -->
            </TableBodyCell>
        </TableBodyRow>
        {/each}
      
    </TableBody>
</Table>

<DeleteModal bind:open={openDeleteModal} bind:dirPath={dirPath} bind:name={deleteFileName} on:refesh_list_msg={()=>dispatch('refesh_list_msg', '')}></DeleteModal>