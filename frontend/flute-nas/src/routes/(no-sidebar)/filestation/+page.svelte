<script lang="ts">
	import MetaTag from '../../../components/MetaTag.svelte';
	import {
		Card,
		Breadcrumb,
		BreadcrumbItem,
		Button,
		Checkbox,
		Heading,
		Indicator
	} from 'flowbite-svelte';
	import { GradientButton } from 'flowbite-svelte';
	import FileTable from '../../../components/files/FileTable.svelte';
	import UploadModal from '../../../components/files/UploadModal.svelte';
	import UploadingToast from '../../../components/files/UploadingToast.svelte';
	import { onMount } from 'svelte';
	import { FluteAPI } from '$lib/api';
	import { FileEntry, FileProgress } from '$lib/model';
	import { RefreshOutline, UploadSolid } from 'flowbite-svelte-icons';
	import { Frame } from 'flowbite-svelte';
	import toast, { Toaster } from 'svelte-french-toast';
	import CreateDirModal from '../../../components/files/CreateDirModal.svelte';
	import Tree from '../../../components/files/Tree.svelte';
	import { type DirTreeNode } from '$lib/interface';
	import { fade } from 'svelte/transition';

	// 左侧 目录导航栏选中的目录
	let selectedPath = '/';
	// 左侧 目录导航栏的数据
	const sampleData: DirTreeNode[] = [
		{
			name: '/',
			path: '/',
			children: []
		}
	];

	let entries: FileEntry[] = [];
	let dirPath: string = '/';
	let loading: boolean = false;
	let group: string[] = [];

	$: readDir(dirPath);
	onMount(() => {
		dirPath = '/';
	});
	// 当左侧目录导航路展开某个目录的时候，右侧文件列表自动定位到该路径
	$: dirPath = selectedPath;
	function readDir(dirPath: string = '/') {
		entries = [];
		loading = true;
		const api = new FluteAPI();
		const originDirPath = dirPath;
		api
			.post('/v1/files/readdir', { Path: dirPath })
			.then((resp) => {
				// console.log(resp.data.Entries)
				entries.length = 0;
				resp.data.Entries.forEach((entry: any) => {
					let e = new FileEntry(entry.Name, entry.IsDir, entry.Size, entry.LastMod, entry.Kind);
					entries.push(e);
					// push 不会触发重新渲染，需要手动赋值一下
					entries = entries;
				});
				// console.log(entries)
				loading = false;
			})
			.catch((err) => {
				dirPath = originDirPath;
				console.log(err);
				loading = false;
			});
	}

	let openUploadModal: boolean = false;
	function toggleUploadModal() {
		openUploadModal = !openUploadModal; // 切换模态框状态
	}

	let openCreateDirModal: boolean = false;
	function toggleCreateDirModal() {
		openCreateDirModal = !openCreateDirModal;
	}
	function toggleDeleteModal() {
		console.log(group);
	}

	let showUploadingToast = false;
	let selectFiles: FileProgress[] = [];
	let speed: string = '';

	// 当上传进度展示Toast关闭时 中断所有的上传任务
	$: if (!showUploadingToast) {
		// Toast关闭时，终止所有上传任务
		if (selectFiles && selectFiles.length > 0) {
			selectFiles.forEach((f) => {
				if (f.xhr) {
					f.xhr.abort();
				}
			});
			selectFiles = [];
		}
	}

	function handleUploading(e: any) {
		// console.log("father receive:");
		// console.log(e.detail);
		if (e.detail) {
			console.log('show uploading toast');
			const api = new FluteAPI();
			showUploadingToast = true;
			let uploadDirPath = dirPath;
			setTimeout(async function () {
				selectFiles = e.detail;
				for (let i = 0; i < selectFiles.length; i++) {
					await api.uploadFile(
						uploadDirPath,
						selectFiles[i],
						// (progressEvent :any) => { // 监听上传进度
						// 	const { loaded, total } = progressEvent;
						// 	if (total !== undefined) {
						// 		const percentCompleted = Math.round((loaded * 100) / total); // 计算百分比
						// 		const elapsedTime = (Date.now() - startTime) / 1000; // 计算已用时间
						// 		speed = formatSpeed(loaded / 1024 / elapsedTime) // 计算网速 (KB/s)

						// 		selectFiles[i].updateProgress(percentCompleted)
						// 		selectFiles = selectFiles
						// 		console.log(`上传进度: ${selectFiles[i].progress}% | 网速: ${speed} KB/s`); // 输出进度和网速
						// 	}
						// }
						(percentCompleted: any, formattedSpeed: any) => {
							selectFiles[i].updateProgress(percentCompleted);
							selectFiles = selectFiles;
							speed = formattedSpeed;
							console.log(`上传进度: ${selectFiles[i].progress}% | 网速: ${speed} KB/s`); // 输出进度和网速
						}
					);
					// 每上传一个文件 刷新一下文件列表 前提是没有切换目录
					if (dirPath == uploadDirPath) {
						readDir(dirPath);
					}
				}
				showUploadingToast = false;
			}, 100);
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
			<BreadcrumbItem href="/filestation">FileStation</BreadcrumbItem>
		</Breadcrumb>
		<Heading tag="h1" class="text-xl font-semibold text-gray-900 sm:text-2xl dark:text-white">
			File Station
		</Heading>
	</div>
	<div class="mt-px space-y-4">
		<div class="grid grid-cols-1 gap-4 xl:grid-cols-3 2xl:grid-cols-3">
			<Card class="justify-start" size="md">
				<div class="items-center justify-between lg:flex">
					<div class="mb-4 mt-px lg:mb-0">
						<Heading tag="h3" class="-ml-0.25 mb-2 text-xl font-semibold dark:text-white">
							Directory Explorer
						</Heading>
						<span class="text-base font-normal text-gray-500 dark:text-gray-400">
							This is a files list of current directory
						</span>
					</div>
				</div>
				<div class="tree-container lg:flex">
					<Tree data={sampleData} bind:selectedPath />
				</div>
			</Card>
			<Card size="xl" class="col-span-2 max-w-none shadow-sm">
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
							<GradientButton color="pinkToOrange" on:click={toggleDeleteModal}
								>Delete</GradientButton
							>
							<Button
								pill={true}
								class="bg-gradient-to-br from-purple-600 to-blue-500 !p-2 text-white shadow-blue-500/50 hover:bg-gradient-to-bl focus:ring-blue-300 dark:shadow-blue-800/80 dark:focus:ring-blue-800"
								on:click={() => {
									readDir(dirPath);
								}}><RefreshOutline class="h-6 w-6" /></Button
							>
							<GradientButton color="purpleToBlue" on:click={toggleUploadModal}
								>Upload</GradientButton
							>
							<GradientButton color="pinkToOrange" on:click={toggleCreateDirModal}
								>CreateDir</GradientButton
							>
						</div>
					</div>
				</div>
				<FileTable
					files={entries}
					bind:loading
					bind:dirPath
					bind:group
					on:refesh_list_msg={() => readDir(dirPath)}
				></FileTable>
			</Card>
		</div>
	</div>
</main>

<!-- Modals -->

<UploadModal bind:open={openUploadModal} on:selected_message={handleUploading}></UploadModal>
<CreateDirModal
	bind:open={openCreateDirModal}
	bind:dirPath
	on:refesh_list_msg={() => readDir(dirPath)}
></CreateDirModal>

<Toaster />

<UploadingToast bind:toastStatus={showUploadingToast} color="indigo">
	<UploadSolid slot="icon" class="h-6 w-6" />
	<span slot="title">Uploading &nbsp;&nbsp;&nbsp;&nbsp;{speed}</span>

	<Frame tag="ul" rounded border class="divide-y divide-gray-200 dark:divide-gray-600">
		{#if selectFiles}
			{#each selectFiles as item, index}
				{#if item.progress < 100 || (item.completedAt && Date.now() - item.completedAt < 2000)}
					<li
						class="w-full list-none truncate px-4 py-2 text-sm font-medium first:rounded-t-lg last:rounded-b-lg"
						style="background-image: linear-gradient(to right, #6ee7b7 0%, #6ee7b7 {item.progress}%, #fff {item.progress2}%, #fff 100%);"
						transition:fade={{ delay: 1000, duration: 500 }}
					>
						{item.file.name}
					</li>
				{/if}
			{/each}
		{/if}
	</Frame>
</UploadingToast>
