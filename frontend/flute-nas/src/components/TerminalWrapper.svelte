<script lang="ts">
    import { onMount, onDestroy, createEventDispatcher } from 'svelte';
    import { FluteAPI } from '$lib/api';
    import { Terminal as XTerm } from '@xterm/xterm';
    import { FitAddon } from '@xterm/addon-fit';
    import { Unicode11Addon } from '@xterm/addon-unicode11';
    import { WebLinksAddon } from '@xterm/addon-web-links';
    import { CurrentHostIP } from '$lib/vars';
    import { TerminalCommandExecutor } from '$lib/TerminalCommandExecutor';
    import "@xterm/xterm/css/xterm.css";

    // Props
    export let terminalId: string = 'terminal';
    export let hostIP: string = '';
    export let terminalName: string = 'terminal';
    export let finderPrint: string = 'unknown';
    export let width: string = '100%';
    export let height: string = '400px';
    export let autoConnect: boolean = true;
    export let initialCommands: string[] = [];

    // Events
    const dispatch = createEventDispatcher<{
        connected: { terminal: any };
        disconnected: {};
        error: { error: any };
        commandStart: { command: string };
        commandComplete: { command: string; exitCode?: number };
        commandFailed: { command: string; exitCode: number };
        allCommandsCompleted: {};
    }>();

    // State
    let terminalElement: HTMLElement | null = null;
    let term: XTerm | null = null;
    let socket: WebSocket | null = null;
    let commandExecutor: TerminalCommandExecutor | null = null;
    let fitAddon: FitAddon | null = null;
    let webLinksAddon: WebLinksAddon | null = null;
    let unicode11Addon: Unicode11Addon | null = null;

    // Internal state
    let isConnected = false;
    let isConnecting = false;

    onMount(() => {
        if (!terminalElement) return;

        // Initialize terminal
        term = new XTerm({
            'allowProposedApi': true,
            rows: 24,
            cols: 80,
            fontSize: 14,
            fontFamily: 'Monaco, Menlo, Ubuntu Mono, monospace',
            cursorBlink: true
        });

        term.open(terminalElement);

        fitAddon = new FitAddon();
        term.loadAddon(fitAddon);
        webLinksAddon = new WebLinksAddon();
        term.loadAddon(webLinksAddon);
        unicode11Addon = new Unicode11Addon();
        term.loadAddon(unicode11Addon);

        // Dispatch terminal instance
        if (term) {
            dispatch('connected', { terminal: term });
        }

        if (autoConnect) {
            connectToTerminal();
        }

        const handleResize = () => {
            if (fitAddon) {
                fitAddon.fit();
            }
        };

        window.addEventListener('resize', handleResize);

        return () => {
            disconnect();
            window.removeEventListener('resize', handleResize);

            if (fitAddon) {
                try {
                    fitAddon.dispose();
                } catch (e) {}
                fitAddon = null;
            }
            if (webLinksAddon) {
                try {
                    webLinksAddon.dispose();
                } catch (e) {}
                webLinksAddon = null;
            }
            if (unicode11Addon) {
                try {
                    unicode11Addon.dispose();
                } catch (e) {}
                unicode11Addon = null;
            }
        };
    });

    async function connectToTerminal() {
        if (isConnecting || isConnected) return;

        isConnecting = true;

        try {
            const textDecoder = new TextDecoder();
            const api = new FluteAPI();
            
            let resolvedHostIP = hostIP || $CurrentHostIP;
            if (!resolvedHostIP) {
                resolvedHostIP = "127.0.0.1";
            }

            const resp = await api.post("/v1/terminal", {
                'HostIP': resolvedHostIP, 
                'FinderPrint': finderPrint, 
                "TerminalName": terminalName
            });

            if (resp.code !== 0) {
                throw new Error('Create terminal connection failed: ' + resp.msg);
            }

            const currentLocation = window.location.host;
            const prefix = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
            const wsUrl = `${prefix}${currentLocation}/ws/v1/terminal?token=${resp.data.Token}`;
            
            socket = new WebSocket(wsUrl);
            socket.binaryType = 'arraybuffer';

            // Initialize the command executor
            if (term) {
                commandExecutor = new TerminalCommandExecutor({
                    term,
                    socket,
                    onCommandStart: (command: string) => {
                        dispatch('commandStart', { command });
                    },
                    onCommandComplete: (command: string, exitCode?: number) => {
                        dispatch('commandComplete', { command, exitCode });
                    },
                    onCommandFailed: (command: string, exitCode: number) => {
                        dispatch('commandFailed', { command, exitCode });
                    },
                    onAllCommandsCompleted: () => {
                        dispatch('allCommandsCompleted', {});
                    },
                    onAllCommandsCompletedAndExit: (shouldCloseWindow: boolean) => {
                        // No window closing for modal component
                    },
                    onError: (error: Error) => {
                        dispatch('error', { error });
                    }
                });
            }

            socket.onopen = function () {
                isConnected = true;
                isConnecting = false;
                term?.writeln('Connected to the server.');
                term?.focus();

                setTimeout(() => {
                    term?.focus();
                    if (fitAddon) {
                        fitAddon.fit();
                    }
                });

                // Execute initial commands if provided
                if (initialCommands && initialCommands.length > 0 && commandExecutor) {
                    commandExecutor.executeCommands(initialCommands);
                } else if (commandExecutor) {
                    commandExecutor.executeStoredCommands();
                }

                // Handle resize events
                term?.onResize(function(event: any) {
                    if (!socket) return;
                    const rows = event.rows;
                    const cols = event.cols;
                    const send = `1:${rows}:${cols}`;
                    socket.send(send);
                });

                // Focus terminal
                term?.focus();
            };

            // Receive backend data and display
            socket.onmessage = function(event) {
                if (!term) return;

                let text;
                if (typeof event.data === 'string') {
                    text = event.data;
                } else {
                    text = textDecoder.decode(new Uint8Array(event.data));
                }

                // Parse OSC 133 sequences to detect command execution status
                const parts = text.split(/(\x1b\]133;\w(?:;\d*)?\x07)/);

                for (let i = 0; i < parts.length; i++) {
                    const part = parts[i];

                    if (part.startsWith('\x1b]133;')) {
                        if (commandExecutor) {
                            commandExecutor.handleOscSequence(part);
                        }
                        continue;
                    }

                    if (part) {
                        term.write(part);
                    }
                }
            };

            socket.onerror = function (event) {
                isConnecting = false;
                dispatch('error', { error: event });
                term?.writeln('WebSocket error. See console for details.');
            };

            socket.onclose = function () {
                isConnected = false;
                isConnecting = false;
                dispatch('disconnected', {});
                term?.writeln('');
                term?.writeln('Disconnected from the server.');

                if (commandExecutor) {
                    const status = commandExecutor.getStatus();
                    if (status.remaining > 0) {
                        term?.writeln(`\x1b[1;31mWarning: ${status.remaining} commands were left in queue due to disconnection.\x1b[0m`);
                    }
                }
            };

            // Send input data
            term?.onData(function(data: string) {
                if (socket) {
                    socket.send("0:" + unescape(encodeURIComponent(data)).length.toString() + ":" + data);
                }
            });

            // Ping-pong keepalive
            const pingInterval = setInterval(function() {
                if (socket && socket.readyState === WebSocket.OPEN) {
                    socket.send("2");
                } else {
                    clearInterval(pingInterval);
                }
            }, 30 * 1000);

        } catch (error) {
            isConnecting = false;
            dispatch('error', { error });
        }
    }

    function disconnect() {
        if (socket) {
            socket.close();
            socket = null;
        }
        isConnected = false;
        isConnecting = false;
    }

    function executeCommands(commands: string[]) {
        if (commandExecutor) {
            commandExecutor.executeCommands(commands);
        }
    }

    function stopExecution() {
        if (commandExecutor) {
            commandExecutor.stopExecution();
        }
    }

    // Expose methods to parent components
    export { executeCommands, stopExecution, connectToTerminal, disconnect };
</script>

<div 
    bind:this={terminalElement}
    id={terminalId} 
    class="terminal-container overflow-hidden"
    style="width: {width}; height: {height};"
/>
