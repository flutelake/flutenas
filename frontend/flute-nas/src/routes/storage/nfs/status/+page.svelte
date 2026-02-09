<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Breadcrumb,
		BreadcrumbItem,
		Button,
		Heading,
		Alert,
		Card,
		Badge,
		P,
		Spinner
	} from 'flowbite-svelte';
	import {
		PlayOutline,
		StopOutline,
		CheckCircleOutline,
		ExclamationCircleOutline,
		CogOutline,
		RefreshOutline,
		TerminalSolid,
		ArrowRightOutline
	} from 'flowbite-svelte-icons';
	import MetaTag from '../../../../components/MetaTag.svelte';
	import { FluteAPI } from '$lib/api';
	import { CurrentHostIP } from '$lib/vars';
	import { goto } from '$app/navigation';
	import TerminalModal from '../../../../components/TerminalModal.svelte';

	const path: string = '/storage/nfs/status';
	const description: string = 'NFS Server Status - FluteNAS Web Console';
	const title: string = 'FluteNAS Web Console - NFS Server Status';
	const subtitle: string = 'NFS Server Control Panel';

	let loading: boolean = false;
	let nfsStatus: string = 'unknown';
	let nfsUptime: string = '';
	let lastError: string = '';
	let successMessage: string = '';
	let validating: boolean = false;
	let validationResult: { valid: boolean; message: string } | null = null;
	let systemInfo: any = null;
	let checkingInstallation: boolean = false;
	let nfsInstalled: boolean = false;
	let distroInfo: string = '';
	let installCommands: string[] = [];

	// Terminal modal state
	let showTerminalModal: boolean = false;
	let terminalRef: any = null;

	// 状态颜色映射
	const statusColors = {
		running: 'green',
		stopped: 'red',
		starting: 'yellow',
		stopping: 'yellow',
		unknown: 'gray',
		not_installed: 'purple'
	};

	// 检查NFS-Ganesha安装状态
	async function checkNFSInstallation() {
		checkingInstallation = true;
		lastError = '';

		try {
			const api = new FluteAPI();
			const hostIP = $CurrentHostIP || '127.0.0.1';

			// 获取系统信息（包含NFS安装状态）
			const response = await api.getHostSystemInfo(hostIP);

			if (response.code === 0) {
				systemInfo = response.data;
				nfsInstalled = response.data.NFSInstalled;
				distroInfo = `${response.data.DistroID} ${response.data.DistroVersion}`;
				installCommands = response.data.InstallCommands || [];

				// 如果已安装，获取服务状态
				if (nfsInstalled) {
					await getNFSStatus();
				} else {
					nfsStatus = 'not_installed';
				}
			} else {
				lastError = response.message || '检查NFS安装状态失败';
			}
		} catch (err: any) {
			console.error('Failed to check NFS installation:', err);
			lastError = err.message || '检查NFS安装状态失败';
		} finally {
			checkingInstallation = false;
		}
	}

	// 获取NFS服务状态
	async function getNFSStatus() {
		loading = true;
		lastError = '';

		try {
			const api = new FluteAPI();
			const hostIP = $CurrentHostIP || '127.0.0.1';

			const response = await api.getNFSServiceStatus(hostIP);

			if (response.code === 0) {
				nfsStatus = response.data.Status;
				nfsUptime = response.data.Uptime || '';
			} else {
				lastError = response.message || '获取NFS服务状态失败';
			}
		} catch (err: any) {
			console.error('Failed to get NFS status:', err);
			lastError = err.message || '获取NFS服务状态失败';
			nfsStatus = 'unknown';
		} finally {
			loading = false;
		}
	}

	// 启动NFS服务
	async function startNFSServer() {
		loading = true;
		lastError = '';
		successMessage = '';
		validationResult = null;

		try {
			const api = new FluteAPI();
			const hostIP = $CurrentHostIP || '127.0.0.1';

			const response = await api.startNFSServer(hostIP);

			if (response.code === 0) {
				successMessage = 'NFS服务已成功启动';
				await getNFSStatus(); // 刷新状态
			} else {
				lastError = response.message || '启动NFS服务失败';
			}
		} catch (err: any) {
			console.error('Failed to start NFS server:', err);
			lastError = err.message || '启动NFS服务失败';
		} finally {
			loading = false;
		}
	}

	// 打开终端安装模态框
	function openTerminalInstall() {
		showTerminalModal = true;
	}

	// 处理终端模态框关闭事件
	function handleTerminalClose() {
		// 终端关闭后刷新页面状态
		checkNFSInstallation();
	}

	// 停止NFS服务
	async function stopNFSServer() {
		loading = true;
		lastError = '';
		successMessage = '';
		validationResult = null;

		try {
			const api = new FluteAPI();
			const hostIP = $CurrentHostIP || '127.0.0.1';

			const response = await api.stopNFSServer(hostIP);

			if (response.code === 0) {
				successMessage = 'NFS服务已成功停止';
				await getNFSStatus(); // 刷新状态
			} else {
				lastError = response.message || '停止NFS服务失败';
			}
		} catch (err: any) {
			console.error('Failed to stop NFS server:', err);
			lastError = err.message || '停止NFS服务失败';
		} finally {
			loading = false;
		}
	}

	// 验证NFS配置
	async function validateNFSConfig() {
		validating = true;
		lastError = '';
		successMessage = '';
		validationResult = null;

		try {
			const api = new FluteAPI();

			const response = await api.post('/v1/nfs-server/validate', {
				ConfigPath: '/etc/ganesha/ganesha.conf'
			});

			if (response.code === 0) {
				validationResult = {
					valid: response.data.Valid,
					message: response.data.Message
				};
			} else {
				lastError = response.message || '验证配置失败';
			}
		} catch (err: any) {
			console.error('Failed to validate NFS config:', err);
			lastError = err.message || '验证配置失败';
		} finally {
			validating = false;
		}
	}

	// 刷新状态
	async function refreshStatus() {
		await checkNFSInstallation();
		successMessage = '状态已刷新';
		setTimeout(() => {
			successMessage = '';
		}, 3000);
	}

	// 获取状态颜色
	function getStatusColor(status: string): any {
		const colors: Record<string, any> = statusColors;
		return colors[status] || 'gray';
	}

	onMount(() => {
		checkNFSInstallation();
	});
</script>

<MetaTag {path} {description} {title} {subtitle} />

<main class="relative h-full w-full overflow-y-auto bg-white dark:bg-gray-800">
	<div class="p-4">
		<Breadcrumb class="mb-5">
			<BreadcrumbItem href="/" home>Home</BreadcrumbItem>
			<BreadcrumbItem href="/storage">Storage</BreadcrumbItem>
			<BreadcrumbItem>NFS</BreadcrumbItem>
			<BreadcrumbItem>Server Status</BreadcrumbItem>
		</Breadcrumb>

		<Heading tag="h1" class="mb-2 text-xl font-semibold text-gray-900 sm:text-2xl dark:text-white">
			NFS Server Control Panel
		</Heading>
		<p class="mb-6 text-lg font-normal text-gray-500 sm:text-xl dark:text-gray-400">
			Monitor and control NFS-Ganesha service status
		</p>

		{#if lastError}
			<Alert color="red" class="mb-4">
				<ExclamationCircleOutline slot="icon" class="h-5 w-5" />
				{lastError}
			</Alert>
		{/if}

		{#if successMessage}
			<Alert color="green" class="mb-4">
				<CheckCircleOutline slot="icon" class="h-5 w-5" />
				{successMessage}
			</Alert>
		{/if}

		<!-- 安装状态提示 -->
		{#if checkingInstallation}
			<Card class="mb-6">
				<div class="flex items-center">
					<Spinner size="4" class="mr-3" />
					<span class="text-gray-700 dark:text-gray-300">检查NFS-Ganesha安装状态...</span>
				</div>
			</Card>
		{/if}

		{#if !nfsInstalled && !checkingInstallation && distroInfo}
			<Alert color="purple" class="mb-4">
				<ExclamationCircleOutline slot="icon" class="h-5 w-5" />
				<div>
					<strong>NFS-Ganesha未安装</strong>
					<p class="mt-1 text-sm">
						检测到系统: {distroInfo}
					</p>
					<div class="mt-3">
						<Button size="sm" color="purple" on:click={openTerminalInstall}>
							<TerminalSolid class="mr-2 h-4 w-4" />
							在终端中安装
							<ArrowRightOutline class="ml-2 h-4 w-4" />
						</Button>
					</div>
				</div>
			</Alert>
		{/if}

		<!-- 状态卡片 -->
		<Card class="mb-6">
			<div class="flex items-center justify-between">
				<div>
					<Heading tag="h3" class="text-lg font-semibold text-gray-900 dark:text-white">
						Service Status
					</Heading>
					<div class="mt-2 flex items-center">
						{#if loading}
							<Spinner size="4" class="mr-2" />
						{/if}
						<Badge color={getStatusColor(nfsStatus)} size="lg" class="text-sm">
							{nfsStatus.toUpperCase()}
						</Badge>
						{#if nfsUptime && nfsStatus === 'running'}
							<span class="ml-3 text-sm text-gray-500 dark:text-gray-400">
								Uptime: {nfsUptime}
							</span>
						{/if}
					</div>
				</div>
				<Button size="sm" color="alternative" on:click={refreshStatus}>
					<RefreshOutline class="mr-2 h-4 w-4" />
					Refresh
				</Button>
			</div>
		</Card>

		<!-- 控制按钮 -->
		<Card class="mb-6">
			<Heading tag="h3" class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
				Service Control
			</Heading>
			<div class="flex flex-wrap gap-3">
				<Button
					color="green"
					on:click={startNFSServer}
					disabled={loading || nfsStatus === 'running' || !nfsInstalled}
					class="min-w-[120px]"
				>
					<PlayOutline class="mr-2 h-5 w-5" />
					Start Service
				</Button>

				<Button
					color="red"
					on:click={stopNFSServer}
					disabled={loading || nfsStatus === 'stopped'}
					class="min-w-[120px]"
				>
					<StopOutline class="mr-2 h-5 w-5" />
					Stop Service
				</Button>

				<Button
					color="blue"
					on:click={validateNFSConfig}
					disabled={validating}
					class="min-w-[140px]"
				>
					{#if validating}
						<Spinner size="4" class="mr-2" />
					{:else}
						<CogOutline class="mr-2 h-5 w-5" />
					{/if}
					Validate Config
				</Button>
			</div>
		</Card>

		<!-- 验证结果 -->
		{#if validationResult}
			<Card class="mb-6">
				<Heading tag="h3" class="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
					Configuration Validation
				</Heading>
				<Alert color={validationResult.valid ? 'green' : 'red'}>
					{#if validationResult.valid}
						<CheckCircleOutline slot="icon" class="h-5 w-5" />
					{:else}
						<ExclamationCircleOutline slot="icon" class="h-5 w-5" />
					{/if}
					{validationResult.message}
				</Alert>
			</Card>
		{/if}

		<!-- 信息卡片 -->
		<Card>
			<Heading tag="h3" class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
				Service Information
			</Heading>
			<div class="grid gap-4 md:grid-cols-2">
				<div class="space-y-2">
					<P class="text-sm text-gray-500 dark:text-gray-400">
						<strong>Service Name:</strong> nfs-ganesha
					</P>
					<P class="text-sm text-gray-500 dark:text-gray-400">
						<strong>Config File:</strong> /etc/ganesha/ganesha.conf
					</P>
				</div>
				<div class="space-y-2">
					<P class="text-sm text-gray-500 dark:text-gray-400">
						<strong>Host:</strong>
						{$CurrentHostIP || '127.0.0.1'}
					</P>
					<P class="text-sm text-gray-500 dark:text-gray-400">
						<strong>Last Check:</strong>
						{new Date().toLocaleString()}
					</P>
				</div>
			</div>
		</Card>
	</div>

	<!-- 终端模态框 -->
	<TerminalModal
		bind:this={terminalRef}
		bind:open={showTerminalModal}
		title="NFS-Ganesha Installation"
		terminalName="nfs-installation-terminal"
		width="800px"
		height="500px"
		initialCommands={installCommands}
		closeOnAllCommandsCompleted={true}
		on:close={handleTerminalClose}
	/>
</main>
