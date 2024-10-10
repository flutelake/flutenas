<script lang="ts">
    import MetaTag from '../../../components/MetaTag.svelte';
    import { Card, Breadcrumb, BreadcrumbItem, Button, Checkbox, Heading, Indicator } from 'flowbite-svelte';
	import { GradientButton } from 'flowbite-svelte';
	import FileTable from '../../../components/files/FileTable.svelte';
	import UploadModal from '../../../components/files/UploadModal.svelte';
	import UploadingToast from '../../../components/files/UploadingToast.svelte';
	import { onMount } from 'svelte';
	import { FluteAPI } from '$lib/api';
	import { FileEntry, FileProgress } from '$lib/model';
	import { formatSpeed } from '$lib/index'
	import {RefreshOutline, UploadSolid} from 'flowbite-svelte-icons'
	import { Frame } from 'flowbite-svelte';
	import toast, { Toaster } from 'svelte-french-toast';
	import CreateDirModal from '../../../components/files/CreateDirModal.svelte';

	let entries :FileEntry[] = []
	let dirPath :string = '/'
	let loading : boolean = false
	$: readDir(dirPath)
	function readDir(dirPath :string = '/') {
		entries = []
		loading = true
		const api = new FluteAPI()
		const originDirPath = dirPath
		api.post("/v1/files/readdir", {"Path": dirPath}).then(resp => {
			// console.log(resp.data.Entries)
			entries.length = 0;
			resp.data.Entries.forEach((entry :any) => {
				let e = new FileEntry(entry.Name, entry.IsDir, entry.Size, entry.LastMod, entry.Kind)
				entries.push(e); 
				// push 不会触发重新渲染，需要手动赋值一下
				entries = entries;
			});
			// console.log(entries)
			loading = false
		}).catch(err => {
			dirPath = originDirPath
			console.log(err)
			loading = false
		})
	}

	let openUploadModal: boolean = false;
	function toggleUploadModal() {
		openUploadModal = !openUploadModal; // 切换模态框状态
	}

	let openCreateDirModal :boolean = false;
	function toggleCreateDirModal() {
		openCreateDirModal = !openCreateDirModal
	}
	onMount(() => {
		readDir(dirPath)
	})

	let showUploadingToast = false
	let selectFiles :FileProgress[] = [];
	let speed :string = ''
	function handleUploading(e :any) {
		// console.log("father receive:");
		// console.log(e.detail);
		if (e.detail) {
			console.log("show uploading toast")
			const api = new FluteAPI()
			showUploadingToast = true
	
			setTimeout(async function(){
				selectFiles = e.detail
				for (let i = 0; i < selectFiles.length; i++) {
					const startTime = Date.now(); // 记录开始时间
					await api.uploadFile(
						dirPath,
						selectFiles[i], 
						(progressEvent :any) => { // 监听上传进度
						const { loaded, total } = progressEvent;
						if (total !== undefined) {
							const percentCompleted = Math.round((loaded * 100) / total); // 计算百分比
							const elapsedTime = (Date.now() - startTime) / 1000; // 计算已用时间
							speed = formatSpeed(loaded / 1024 / elapsedTime) // 计算网速 (KB/s)

							selectFiles[i].updateProgress(percentCompleted)
							selectFiles = selectFiles
							console.log(`上传进度: ${selectFiles[i].progress}% | 网速: ${speed} KB/s`); // 输出进度和网速
						}
                	})
					// 每上传一个文件 刷新一下文件列表
					readDir(dirPath)
				}
				showUploadingToast = false
				
			}, 100)
		}
	}
	
	const path: string = '/filestation';
    const description: string = 'FileStation - flute nas console';
	const metaTitle: string = 'FluteNAS Web Console - FileStation';
    const subtitle: string = 'file station';
</script>

<MetaTag {path} {description} title={metaTitle} {subtitle} />

<main class="relative h-full w-full overflow-y-auto bg-white dark:bg-gray-800">
	<div class="p-4">
		<Breadcrumb class="mb-5">
			<BreadcrumbItem home>Home</BreadcrumbItem>
			<BreadcrumbItem href="/crud/users">FileStation</BreadcrumbItem>
		</Breadcrumb>
		<Heading tag="h1" class="text-xl font-semibold text-gray-900 dark:text-white sm:text-2xl">
			File Station
		</Heading>
	</div>
	<div class="mt-px space-y-4">
		<div class="grid grid-cols-1 gap-4 xl:grid-cols-3 2xl:grid-cols-3">
			<Card horizontal class="items-center justify-between" size="md">
				<div class="w-full">
					<p>Directory Explorer</p>
				</div>
			</Card>
			<Card size="xl" class="shadow-sm max-w-none col-span-2">
				<div class="items-center justify-between lg:flex">
					<div class="mb-4 mt-px lg:mb-0">
						<Heading tag="h3" class="-ml-0.25 mb-2 text-xl font-semibold dark:text-white">
							File Explorer
						</Heading>
						<span class="text-base font-normal text-gray-500 dark:text-gray-400">
							This is a files list of current directory
						</span>
					</div>
					<div class="items-center justify-between gap-3 space-y-4 sm:flex sm:space-y-0">
						<div class="flex items-center">
							<input placeholder="/" bind:value={dirPath} class="border" />
						</div>
						<div class="flex items-center space-x-4">
							<Button pill={true} class="!p-2 text-white bg-gradient-to-br from-purple-600 to-blue-500 hover:bg-gradient-to-bl focus:ring-blue-300 dark:focus:ring-blue-800 shadow-blue-500/50 dark:shadow-blue-800/80" on:click={() => {readDir(dirPath)}}><RefreshOutline class="w-6 h-6" /></Button>
							<GradientButton color="purpleToBlue" on:click={toggleUploadModal}>Upload</GradientButton>
							<GradientButton color="pinkToOrange" on:click={toggleCreateDirModal}>CreateDir</GradientButton>
						</div>
					</div>
				</div>
				<FileTable files={entries} bind:loading={loading} bind:dirPath={dirPath} on:refesh_list_msg={()=>readDir(dirPath)}></FileTable>
			</Card>
		</div>
	</div>
</main>

<!-- Modals -->

<UploadModal bind:open={openUploadModal} on:selected_message={handleUploading}></UploadModal>
<CreateDirModal bind:open={openCreateDirModal} bind:dirPath={dirPath} on:refesh_list_msg={()=>readDir(dirPath)}></CreateDirModal>

<Toaster />

<UploadingToast bind:toastStatus={showUploadingToast} color="indigo">
	<UploadSolid slot="icon" class="w-6 h-6" />
	<span slot="title">Uploading &nbsp;&nbsp;&nbsp;&nbsp;{speed}</span>

	<Frame tag='ul' rounded border class="divide-y divide-gray-200 dark:divide-gray-600">
		{#if selectFiles}
            {#each selectFiles as item, index}
            <li
            class="py-2 px-4 w-full text-sm font-medium list-none first:rounded-t-lg last:rounded-b-lg truncate"
            style="background-image: linear-gradient(to right, #6ee7b7 0%, #6ee7b7 {item.progress}%, #fff {item.progress2}%, #fff 100%);"
            >{item.file.name}</li>
            {/each}
        {/if}
	</Frame>
</UploadingToast >
