<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import TerminalWrapper from './TerminalWrapper.svelte';
	import { Modal } from 'flowbite-svelte';

	// Props
	export let open: boolean = false;
	export let title: string = 'Terminal';
	export let terminalId: string = 'terminal-modal';
	export let hostIP: string = '';
	export let terminalName: string = 'terminal-modal';
	export let finderPrint: string = 'unknown';
	export let width: string = '800px';
	export let height: string = '500px';
	export let initialCommands: string[] = [];
	export let showCloseButton: boolean = true;
	export let closeOnEscape: boolean = true;
	export let closeOnOverlayClick: boolean = true;
	export let closeOnAllCommandsCompleted: boolean = false; // Auto-close after commands complete by default

	// Events
	const dispatch = createEventDispatcher<{
		close: {};
		open: {};
		connected: { terminal: any };
		disconnected: {};
		error: { error: any };
		commandStart: { command: string };
		commandComplete: { command: string; exitCode?: number };
		commandFailed: { command: string; exitCode: number };
		allCommandsCompleted: {};
	}>();

	// Local state
	let terminalRef: any = null;

	// Prevent background scrolling when modal is open
	$: if (open) {
		document.body.classList.add('modal-open');
	} else {
		document.body.classList.remove('modal-open');
	}

	function handleClose() {
		open = false;
		// Remove class when modal closes
		document.body.classList.remove('modal-open');
		dispatch('close', {});
	}

	function handleOpen() {
		dispatch('open', {});
	}

	function handleTerminalConnected(e: CustomEvent<{ terminal: any }>) {
		dispatch('connected', e.detail);
	}

	function handleTerminalDisconnected() {
		dispatch('disconnected', {});
	}

	function handleTerminalError(e: CustomEvent<{ error: any }>) {
		dispatch('error', e.detail);
	}

	function handleCommandStart(e: CustomEvent<{ command: string }>) {
		dispatch('commandStart', e.detail);
	}

	function handleCommandComplete(e: CustomEvent<{ command: string; exitCode?: number }>) {
		dispatch('commandComplete', e.detail);
	}

	function handleCommandFailed(e: CustomEvent<{ command: string; exitCode: number }>) {
		dispatch('commandFailed', e.detail);
	}

	function handleAllCommandsCompleted() {
		dispatch('allCommandsCompleted', {});
		// Auto-close the modal if the prop is enabled
		if (closeOnAllCommandsCompleted) {
			handleClose();
		}
	}

	// Methods to expose to parent
	function executeCommands(commands: string[]) {
		if (terminalRef) {
			terminalRef.executeCommands(commands);
		}
	}

	function stopExecution() {
		if (terminalRef) {
			terminalRef.stopExecution();
		}
	}

	function connectToTerminal() {
		if (terminalRef) {
			return terminalRef.connectToTerminal();
		}
	}

	function disconnect() {
		if (terminalRef) {
			terminalRef.disconnect();
		}
	}

	export { executeCommands, stopExecution, connectToTerminal, disconnect };
</script>

<Modal
	size="xl"
	autoclose={closeOnEscape || closeOnOverlayClick}
	bind:open
	on:close={handleClose}
	on:open={handleOpen}
>
	<div slot="header" class="flex items-center justify-between">
		<h3 class="text-lg font-medium text-gray-900 dark:text-white">{title}</h3>
		{#if showCloseButton}
			<button
				type="button"
				class="ml-auto inline-flex items-center rounded-lg bg-transparent p-1.5 text-sm text-gray-400 hover:bg-gray-200 hover:text-gray-900 dark:hover:bg-gray-600 dark:hover:text-white"
				on:click={handleClose}
			>
				<svg
					class="h-5 w-5"
					fill="currentColor"
					viewBox="0 0 20 20"
					xmlns="http://www.w3.org/2000/svg"
				>
					<path
						fill-rule="evenodd"
						d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
						clip-rule="evenodd"
					></path>
				</svg>
			</button>
		{/if}
	</div>
	<div slot="default" class="m-0 max-h-[70vh] p-0">
		<TerminalWrapper
			bind:this={terminalRef}
			{terminalId}
			{hostIP}
			{terminalName}
			{finderPrint}
			width="100%"
			height="100%"
			{initialCommands}
			on:connected={handleTerminalConnected}
			on:disconnected={handleTerminalDisconnected}
			on:error={handleTerminalError}
			on:commandStart={handleCommandStart}
			on:commandComplete={handleCommandComplete}
			on:commandFailed={handleCommandFailed}
			on:allCommandsCompleted={handleAllCommandsCompleted}
		/>
	</div>
</Modal>
