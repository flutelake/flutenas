<script lang="ts">
	import ChartWidget from '../widgets/ChartWidget.svelte';
	import { Card, Chart, Spinner } from 'flowbite-svelte';
	import type { ApexOptions } from 'apexcharts';
	import { onMount } from 'svelte';
	import chart_options_func from '../../routes/(no-sidebar)/overview/chart_options';
	import { FluteAPI } from '$lib/api';
	import type {
		HostMonitoringResponse,
		VictoriaMetricSeries,
		VictoriaQueryRangeResponse
	} from '$lib/model';
	import { CurrentHostIP } from '$lib/vars';
	import { get } from 'svelte/store';

	let chartOptions = chart_options_func(false) as ApexOptions;

	let dark = false;

	type MonitoringPoint = {
		timestamp: number;
		cpu: number;
		mem: number;
		root: number;
	};

	const maxHistoryPoints = 30;

	let history: MonitoringPoint[] = [];

	type DiskHistoryPoint = {
		timestamp: number;
		usage: number;
		isHDD: boolean;
		mountPoint: string;
		device: string;
	};

	type DiskMeta = {
		isHDD: boolean;
		mountPoint: string;
		device: string;
		filesystem: string;
	};

	let diskHistory: Record<string, DiskHistoryPoint[]> = {};
	let diskMeta: Record<string, DiskMeta> = {};

	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let vmPollTimer: ReturnType<typeof setInterval> | null = null;

	let monitoringLoading = false;
	let monitoringError = '';
	let monitoring: HostMonitoringResponse | null = null;

	let vmHistoryLoaded = false;
	let vmHistoryError = '';

	const api = new FluteAPI();

	const timeRangeOptions = {
		'Last 1 hour': 3600,
		'Last 6 hours': 6 * 3600,
		'Last 24 hours': 24 * 3600
	} as const;

	type TimeRangeLabel = keyof typeof timeRangeOptions;
	let selectedTimeRange: TimeRangeLabel = 'Last 1 hour';
	let currentHostIP = '';
	let historyRangeSeconds: number = timeRangeOptions[selectedTimeRange];

	$: historyRangeSeconds = timeRangeOptions[selectedTimeRange];

	const metricQueries = {
		cpu: (hostIP: string) => `flutenas_node_cpu_usage_percent{host="${hostIP}"}`,
		mem: (hostIP: string) => `flutenas_node_mem_usage_percent{host="${hostIP}"}`,
		root: (hostIP: string) => `flutenas_node_root_usage_percent{host="${hostIP}"}`,
		diskUsage: (hostIP: string) => `flutenas_data_disk_usage_percent{host="${hostIP}"}`
	} as const;

	async function fetchVictoriaSeries(
		query: string,
		durationSeconds: number
	): Promise<VictoriaMetricSeries[]> {
		const end = Math.floor(Date.now() / 1000);
		const start = end - durationSeconds;
		const body = {
			Query: query,
			Start: start,
			End: end,
			Step: 10
		};
		const resp: VictoriaQueryRangeResponse = await api.queryVictoriaMetricsRange(body);
		if (!resp || resp.status !== 'success' || !resp.data) {
			throw new Error('VictoriaMetrics query failed');
		}
		return resp.data.result || [];
	}

	async function loadHistoryFromVictoria(hostIP: string, durationSeconds: number) {
		vmHistoryError = '';
		try {
			const [cpuSeriesList, memSeriesList, rootSeriesList, diskSeriesList] = await Promise.all([
				fetchVictoriaSeries(metricQueries.cpu(hostIP), durationSeconds),
				fetchVictoriaSeries(metricQueries.mem(hostIP), durationSeconds),
				fetchVictoriaSeries(metricQueries.root(hostIP), durationSeconds),
				fetchVictoriaSeries(metricQueries.diskUsage(hostIP), durationSeconds)
			]);

			const cpuSeries = cpuSeriesList[0];
			const memSeries = memSeriesList[0];
			const rootSeries = rootSeriesList[0];

			if (cpuSeries && memSeries && rootSeries) {
				const nextHistory: MonitoringPoint[] = [];
				const cpuValues = cpuSeries.values || [];
				const memValues = memSeries.values || [];
				const rootValues = rootSeries.values || [];

				for (let i = 0; i < cpuValues.length; i += 1) {
					const [ts, cpuStr] = cpuValues[i];
					const memPair = memValues[i] || memValues[memValues.length - 1] || [ts, '0'];
					const rootPair = rootValues[i] || rootValues[rootValues.length - 1] || [ts, '0'];
					const timestamp = Math.floor(ts * 1000);
					nextHistory.push({
						timestamp,
						cpu: parseFloat(cpuStr),
						mem: parseFloat(memPair[1]),
						root: parseFloat(rootPair[1])
					});
				}

				history = nextHistory;
				rebuildChartOptions(dark);
			}

			const nextDiskHistory: Record<string, DiskHistoryPoint[]> = {};
			const nextDiskMeta: Record<string, DiskMeta> = {};
			for (const series of diskSeriesList) {
				const labels = series.metric || {};
				const key = labels.device || labels.mount_point;
				if (!key) {
					continue;
				}
				const isHDD = labels.hdd === 'true';
				const mountPoint = labels.mount_point || '';
				const device = labels.device || '';
				const filesystem = labels.filesystem || '';
				const points: DiskHistoryPoint[] = [];
				for (const [ts, vStr] of series.values || []) {
					const timestamp = Math.floor(ts * 1000);
					const usage = parseFloat(vStr);
					points.push({
						timestamp,
						usage,
						isHDD,
						mountPoint,
						device
					});
				}
				const sliced =
					points.length > maxHistoryPoints
						? points.slice(points.length - maxHistoryPoints)
						: points;
				nextDiskHistory[key] = sliced;
				nextDiskMeta[key] = {
					isHDD,
					mountPoint,
					device,
					filesystem
				};
			}
			if (Object.keys(nextDiskHistory).length > 0) {
				diskHistory = nextDiskHistory;
				diskMeta = nextDiskMeta;
			}

			vmHistoryLoaded = true;
		} catch (err) {
			console.error('Failed to load history from VictoriaMetrics', err);
			vmHistoryError = '无法从 VictoriaMetrics 加载历史数据';
		}
	}

	function rebuildChartOptions(isDark: boolean) {
		const baseOptions = chart_options_func(isDark) as ApexOptions;
		const categories = history.map((p) => new Date(p.timestamp).toISOString());

		baseOptions.xaxis = {
			...(baseOptions.xaxis || {}),
			type: 'datetime',
			categories
		};

		baseOptions.series = [
			{
				name: 'CPU',
				data: history.map((p) => Number(p.cpu.toFixed(1)))
			},
			{
				name: '内存',
				data: history.map((p) => Number(p.mem.toFixed(1)))
			},
			{
				name: '根分区',
				data: history.map((p) => Number(p.root.toFixed(1)))
			}
		];

		chartOptions = baseOptions;
	}

	function buildDiskChartOptions(key: string): ApexOptions {
		const points = diskHistory[key] || [];
		const baseOptions = chart_options_func(dark) as ApexOptions;
		const categories = points.map((p) => new Date(p.timestamp).toISOString());

		baseOptions.xaxis = {
			...(baseOptions.xaxis || {}),
			type: 'datetime',
			categories
		};

		baseOptions.series = [
			{
				name: '使用率',
				data: points.map((p) => Number(p.usage.toFixed(1)))
			}
		];

		const meta = diskMeta[key];
		const color = meta && meta.isHDD ? '#f97316' : '#22c55e';
		baseOptions.colors = [color];

		return baseOptions;
	}

	function formatMountLabel(mp: string): string {
		if (!mp) {
			return '';
		}
		if (mp.startsWith('/mnt')) {
			const trimmed = mp.slice(4);
			return trimmed || '/';
		}
		return mp;
	}

	function handler(ev: Event) {
		if ('detail' in ev) {
			dark = !!ev.detail;
			rebuildChartOptions(dark);
		}
	}

	$: if (currentHostIP && historyRangeSeconds) {
		loadHistoryFromVictoria(currentHostIP, historyRangeSeconds);
	}

	onMount(() => {
		document.addEventListener('dark', handler);
		const ip = get(CurrentHostIP) || '127.0.0.1';
		currentHostIP = ip;
		loadMonitoring(ip);
		loadHistoryFromVictoria(ip, historyRangeSeconds);
		pollTimer = setInterval(() => {
			refreshMonitoring(ip);
		}, 10000);
		vmPollTimer = setInterval(() => {
			if (currentHostIP && historyRangeSeconds) {
				loadHistoryFromVictoria(currentHostIP, historyRangeSeconds);
			}
		}, 10000);
		return () => {
			document.removeEventListener('dark', handler);
			if (pollTimer) {
				clearInterval(pollTimer);
				pollTimer = null;
			}
			if (vmPollTimer) {
				clearInterval(vmPollTimer);
				vmPollTimer = null;
			}
		};
	});

	async function loadMonitoring(hostIP: string) {
		monitoringLoading = true;
		monitoringError = '';
		try {
			const result = await api.getHostMonitoringMetrics(hostIP);
			monitoring = result;
		} catch (err) {
			console.error('Failed to load monitoring metrics', err);
			monitoringError = '无法加载监控数据';
		} finally {
			monitoringLoading = false;
		}
	}

	async function refreshMonitoring(hostIP: string) {
		try {
			const result = await api.getHostMonitoringMetrics(hostIP);
			monitoring = result;
		} catch (err) {
			console.error('Failed to refresh monitoring metrics', err);
		}
	}

	function formatPercent(v: number | undefined): string {
		if (v === undefined || isNaN(v)) {
			return '-';
		}
		return v.toFixed(1) + '%';
	}
</script>

<div class="mt-px space-y-4">
	<div class="grid gap-4 xl:grid-cols-2 2xl:grid-cols-3">
		<Card class="h-full" size="xl">
			<div
				class="flex items-center justify-between border-b border-gray-200 pb-4 dark:border-gray-700"
			>
				<div>
					<p class="text-sm font-medium text-gray-500 dark:text-gray-400">节点监控</p>
					<p class="text-lg font-semibold text-gray-900 dark:text-white">
						{monitoring ? monitoring.HostIP : '当前节点'}
					</p>
					{#if monitoring?.Timestamp}
						<p class="mt-1 text-xs text-gray-400">
							采集时间: {monitoring.Timestamp}
						</p>
					{/if}
				</div>
				{#if monitoringLoading}
					<Spinner size="md" />
				{:else if monitoringError}
					<span class="text-xs text-red-500">{monitoringError}</span>
				{/if}
			</div>
			<div class="mt-4 grid grid-cols-2 gap-4">
				<div>
					<p class="mb-1 text-xs text-gray-500 dark:text-gray-400">CPU 使用率</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-white">
						{monitoring ? formatPercent(monitoring.Node.CPUUsagePercent) : '--'}
					</p>
				</div>
				<div>
					<p class="mb-1 text-xs text-gray-500 dark:text-gray-400">内存使用率</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-white">
						{monitoring ? formatPercent(monitoring.Node.MemUsagePercent) : '--'}
					</p>
				</div>
				<div>
					<p class="mb-1 text-xs text-gray-500 dark:text-gray-400">根分区使用率</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-white">
						{monitoring ? formatPercent(monitoring.Node.RootUsagePercent) : '--'}
					</p>
				</div>
				<div class="space-y-1">
					<p class="mb-1 text-xs text-gray-500 dark:text-gray-400">服务状态</p>
					<p class="text-xs text-gray-600 dark:text-gray-300">
						Samba:
						<span class="font-semibold">
							{monitoring ? monitoring.Samba.Status : '--'}
						</span>
						{#if monitoring}
							<span class="ml-1 text-gray-400">
								({monitoring.Samba.ActiveConnections} 连接)
							</span>
						{/if}
					</p>
					<p class="text-xs text-gray-600 dark:text-gray-300">
						NFS:
						<span class="font-semibold">
							{monitoring ? monitoring.NFS.Status : '--'}
						</span>
						{#if monitoring}
							<span class="ml-1 text-gray-400">
								({monitoring.NFS.ActiveConnections} 连接)
							</span>
						{/if}
					</p>
				</div>
			</div>
		</Card>

		<ChartWidget
			{chartOptions}
			title="节点资源使用趋势"
			subtitle="CPU / 内存 / 根分区 使用率"
			bind:timeslot={selectedTimeRange}
			timeslots={timeRangeOptions}
		/>
	</div>

	{#if Object.keys(diskHistory).length > 0}
		<div class="space-y-3">
			<p class="text-sm font-medium text-gray-500 dark:text-gray-400">/mnt 磁盘使用趋势</p>
			<div class="grid gap-4 xl:grid-cols-2 2xl:grid-cols-3">
				{#each Object.keys(diskHistory) as key}
					{#if diskHistory[key] && diskHistory[key].length > 0}
						<ChartWidget
							chartOptions={buildDiskChartOptions(key)}
							title={formatMountLabel(diskMeta[key]?.mountPoint || diskMeta[key]?.device || key)}
							subtitle={diskMeta[key]?.isHDD ? 'HDD' : 'SSD'}
						/>
					{/if}
				{/each}
			</div>
		</div>
	{/if}
</div>
