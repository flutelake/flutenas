<script lang="ts">
	import { FluteAPI } from '$lib/api';
	import { type DirTreeNode } from '$lib/interface';
	import { onMount } from 'svelte';

	export let data: DirTreeNode[] = [];
	export let selectedPath = '';
	/*
	 * 当前Tree.svelte的实现使用selectedPath来控制展开状态，点击时会设置selectedPath为当前项的path。要实现点击展开后再次点击可以折叠，需要：
	 * 添加一个状态变量来跟踪每个节点的展开/折叠状态
	 * 修改点击处理逻辑，当点击已展开的节点时将其折叠
	 * 更新渲染逻辑以使用新的展开状态
	 */
	export let expandedPaths: Set<string> = new Set();

	$: readChild(selectedPath);
	onMount(() => {
		// onMount中通过赋值的方式 出发readChild更新，而不是直接通过调用readChild方法
		// 因为$: 会监听selectedPath的变化，会导致重复调用
		if (selectedPath == '') {
			selectedPath = '/';
		}
	});
	function readChild(dirPath: string = '/') {
		if (dirPath == '') {
			return;
		}
		console.log('readChild: ', dirPath);
		const api = new FluteAPI();
		const originDirPath = dirPath;
		api
			.post('/v1/files/listdir', { Path: dirPath })
			.then((resp) => {
				// find current item
				let currentIndex = -1;
				for (let i = 0; i < data.length; i++) {
					if (data[i].path === dirPath) {
						currentIndex = i;
						break;
					}
				}
				if (currentIndex == -1) {
					console.log('Error: cannot find current item');
					return;
				}
				if (data[currentIndex].children == undefined) {
					data[currentIndex].children = [];
				}
				if (resp.data.Dirs.length == 0) {
					return;
				}
				let change = false;
				resp.data.Dirs.forEach((e: string) => {
					if (e != undefined && e != '') {
						console.log(e);
						let l = data[currentIndex].children?.length;
						if (l == undefined) {
							l = 0;
						}
						// 判断是否重复
						let exist = false;
						for (let i = 0; i < l; i++) {
							if (data[currentIndex].children?.[i].name == e) {
								exist = true;
							}
						}
						if (!exist) {
							data[currentIndex].children?.push({
								name: e,
								path: dirPath + '/' + e,
								children: []
							});

							change = true;
						}
					}
					if (change) {
						// push 不会触发重新渲染，需要手动赋值一下
						data = data;
					}
				});
				if (dirPath == '/') {
					// console.log("is root dir, data: ", data)
					// 默认展开/目录
					expandedPaths.add(dirPath);
					expandedPaths = expandedPaths; // 触发刷新
				}
			})
			.catch((err) => {
				dirPath = originDirPath;
				console.log(err);
			});
	}

	function handleSelect(path: string) {
		console.log('expandedPaths:', expandedPaths);
		if (expandedPaths.has(path)) {
			expandedPaths.delete(path);
			expandedPaths = expandedPaths; // 触发刷新
		} else {
			expandedPaths.add(path);
			expandedPaths = expandedPaths; // 触发刷新
			selectedPath = path;
		}
	}
</script>

<div>
	{#each data as item}
		<button
			style="display: block;"
			class="tree-node {selectedPath === item.path ? 'selected' : ''}"
			on:click={() => handleSelect(item.path)}
		>
			{#if item.children}
				{#if expandedPaths.has(item.path)}
					▽ {item.name}
				{:else}
					▶ {item.name}
				{/if}
			{:else}
				• {item.name}
			{/if}
		</button>

		{#if item.children && expandedPaths.has(item.path)}
			<div class="tree-node-child">
				<svelte:self data={item.children} bind:selectedPath />
			</div>
		{/if}
	{/each}
</div>

<style>
	.tree-node {
		padding-left: 1rem;
		cursor: pointer;
	}
	.tree-node-child {
		padding-left: 2rem;
	}
	.selected {
		background-color: #f0f0f0;
	}
</style>
