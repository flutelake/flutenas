<script lang="ts">
	import { onMount } from 'svelte';
	import MetaTag from '../../../components/MetaTag.svelte';
	import { FluteAPI } from '$lib/api';
	import {
		Avatar,
		Breadcrumb,
		BreadcrumbItem,
		Button,
		Checkbox,
		Heading,
		Indicator
	} from 'flowbite-svelte';
	import { Terminal as XTerm } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { Unicode11Addon } from '@xterm/addon-unicode11';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import { CurrentHostIP } from '$lib/vars';
	import '@xterm/xterm/css/xterm.css';

	const path: string = '/terminal';
	const description: string = 'Terminal - flute nas console';
	const metaTitle: string = 'FluteNAS Web Console - Terminal';
	const subtitle: string = 'terminal';

	onMount(() => {
		const textDecoder = new TextDecoder();
		const term = new XTerm({ allowProposedApi: true });
		const terminalElement = document.getElementById('terminal');
		if (terminalElement) {
			term.open(terminalElement);
		}

		const fitAddon = new FitAddon();
		term.loadAddon(fitAddon);
		term.loadAddon(new WebLinksAddon());
		term.loadAddon(new Unicode11Addon());

		term.writeln('Connecting to server...');
		var api = new FluteAPI();
		console.log(0);
		let hostIP = $CurrentHostIP;
		if (hostIP == '') {
			hostIP = '127.0.0.1';
		}
		api
			.post('/v1/terminal', { HostIP: hostIP, FinderPrint: 'unknown', TerminalName: '01' })
			.then((resp) => {
				// console.log(resp)
				if (resp.code != 0) {
					term.writeln('Create terminal install failed.');
					return;
				}
				console.log(resp);
				let currentLocation = window.location.host;
				let prefix = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
				const socket = new WebSocket(
					prefix + currentLocation + '/ws/v1/terminal?token=' + resp.data.Token
				);
				socket.binaryType = 'arraybuffer';

				socket.onopen = function () {
					term.writeln('Connected to the server.');
					// term._initialized = true;
					term.focus();
					setTimeout(function () {
						term.focus();
						fitAddon.fit();
					});
					// send window size info
					term.onResize(function (event: any) {
						var rows = event.rows;
						var cols = event.cols;
						var send = '1:' + rows + ':' + cols;
						// console.log('resizing to', size);
						socket.send(send);
					});
					window.onresize = function () {
						fitAddon.fit();
					};
					term.focus();
				};

				// receive backend data and display
				socket.onmessage = function (event) {
					if (typeof event.data === 'string') {
						term.write(event.data);
					} else {
						const text = textDecoder.decode(new Uint8Array(event.data));
						term.write(text);
					}
				};

				socket.onerror = function (event) {
					console.error('WebSocket error:', event);
					term.writeln('WebSocket error. See console for details.');
				};

				socket.onclose = function () {
					term.writeln('');
					term.writeln('Disconnected from the server.');
				};

				// send input data
				term.onData(function (data: string) {
					socket.send('0:' + unescape(encodeURIComponent(data)).length.toString() + ':' + data);
				});

				// ping - pong
				setInterval(function () {
					socket.send('2');
				}, 30 * 1000);
			})
			.catch((err) => {
				console.log(err);
			});
	});
</script>

<MetaTag {path} {description} title={metaTitle} {subtitle} />

<main class="relative h-full w-full overflow-y-auto bg-white dark:bg-gray-800">
	<div class="p-4">
		<Breadcrumb class="mb-5">
			<BreadcrumbItem home>Home</BreadcrumbItem>
			<BreadcrumbItem href="/terminal">Terminal</BreadcrumbItem>
		</Breadcrumb>
		<Heading tag="h1" class="text-xl font-semibold text-gray-900 sm:text-2xl dark:text-white">
			Terminal
		</Heading>
	</div>

	<div id="terminal" class="h-full w-full"></div>
</main>

<!-- Modals -->

<!-- <User bind:open={openUser} data={current_user} />
<Delete bind:open={openDelete} /> -->
