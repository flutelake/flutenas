export interface CommandExecutionOptions {
	title?: string;
	commands: string[];
}

export interface TerminalCommandExecutorOptions {
	term: any; // XTerm instance
	socket: WebSocket;
	onCommandStart?: (command: string) => void;
	onCommandComplete?: (command: string, exitCode?: number) => void;
	onCommandFailed?: (command: string, exitCode: number) => void;
	onAllCommandsCompleted?: () => void;
	onAllCommandsCompletedAndExit?: (shouldCloseWindow: boolean) => void; // 新增回调函数
	onError?: (error: Error) => void;
}

export class TerminalCommandExecutor {
	private commandExecuting = false;
	private commandQueue: string[] = [];
	private options: TerminalCommandExecutorOptions;
	private lastExecutedCommand: string | null = null;
	private waitingForPrompt = false; // 新增：标记是否等待提示符，等待连接建立完成

	constructor(options: TerminalCommandExecutorOptions) {
		this.options = options;
	}

	/**
	 * Execute commands automatically from storage or provided options
	 */
	public executeStoredCommands(): void {
		// Check for commands to auto-execute from both localStorage and sessionStorage
		let commandsJson: string | null = localStorage.getItem('terminal_commands');
		let title: string | null = localStorage.getItem('terminal_title');

		// Fallback to sessionStorage if not found in localStorage
		if (!commandsJson) {
			commandsJson = sessionStorage.getItem('terminal_commands');
			title = sessionStorage.getItem('terminal_title');
		}

		if (title) {
			this.options.term.writeln(`\x1b[1;34m${title}\x1b[0m`);
			this.options.term.writeln('='.repeat(title.length));
			this.options.term.writeln('');
		}

		if (commandsJson) {
			try {
				const commands: string[] = JSON.parse(commandsJson) as string[];
				if (commands && commands.length > 0) {
					this.options.term.writeln('\x1b[1;33mAuto-executing installation commands:\x1b[0m');

					// Clear the commands from storage after reading (try both localStorage and sessionStorage)
					if (localStorage.getItem('terminal_commands')) {
						localStorage.removeItem('terminal_commands');
						localStorage.removeItem('terminal_title');
					}
					if (sessionStorage.getItem('terminal_commands')) {
						sessionStorage.removeItem('terminal_commands');
						sessionStorage.removeItem('terminal_title');
					}

					this.executeCommands(commands);
				}
			} catch (error) {
				console.error('Error parsing commands from storage:', error);
				this.options.term.writeln('\x1b[1;31mError parsing installation commands.\x1b[0m');
				if (this.options.onError) {
					this.options.onError(error as Error);
				}
			}
		}
	}

	/**
	 * Execute a set of commands in sequence
	 */
	public executeCommands(commands: string[]): void {
		if (commands.length === 0) return;

		// Add commands to the execution queue
		this.commandQueue = [...commands];
		this.totalCommands = commands.length;

		// 不再立即执行第一个命令，而是等待133;A信号
		this.options.term.writeln('Waiting for terminal ready signal...');
	}

	private processNextCommand = (): void => {
		if (this.commandQueue.length > 0 && !this.commandExecuting) {
			const nextCommand = this.commandQueue.shift();
			if (nextCommand) {
				this.options.term.writeln(`\x1b[1;37mExecuting:\x1b[0m ${nextCommand}`);

				if (this.options.onCommandStart) {
					this.options.onCommandStart(nextCommand);
				}

				this.lastExecutedCommand = nextCommand;
				this.options.socket.send(
					'0:' +
						unescape(encodeURIComponent(nextCommand + '\n')).length.toString() +
						':' +
						nextCommand +
						'\n'
				);
				this.commandExecuting = true; // Mark that a command is now executing
			}
		}
	};

	/**
	 * Handle command completion based on OSC 133 sequences
	 */
	public handleOscSequence(sequence: string): void {
		if (sequence.includes('133;A')) {
			if (this.waitingForPrompt) {
				this.waitingForPrompt = false;
			}
		} else if (sequence.includes('133;D')) {
			if (this.waitingForPrompt) {
				return;
			}

			// Command ended - extract exit status if available
			this.commandExecuting = false;

			// Extract exit status from the sequence: \x1b]133;D;exit_code\x07
			const exitStatusMatch = sequence.match(/\x1b\]133;D;(\d+)\x07/);
			let exitStatus = 0;
			if (exitStatusMatch && exitStatusMatch[1]) {
				exitStatus = parseInt(exitStatusMatch[1], 10);
			}

			if (exitStatus !== 0) {
				// Command failed - stop further execution
				this.options.term.writeln(
					`\x1b[1;31mCommand failed with exit status: ${exitStatus}. Stopping further execution.\x1b[0m`
				);
				this.commandQueue = []; // Clear the remaining command queue

				if (this.options.onCommandFailed && this.lastExecutedCommand) {
					this.options.onCommandFailed(this.lastExecutedCommand, exitStatus);
				}
			} else {
				// Command succeeded - process next command in queue if available
				if (this.commandQueue.length > 0) {
					setTimeout(this.processNextCommand, 100); // Small delay to ensure terminal is ready
				} else {
					if (this.options.onCommandComplete) {
						if (this.lastExecutedCommand) {
							this.options.onCommandComplete(this.lastExecutedCommand, exitStatus);
						}
					}
					// Check if all commands are completed
					this.checkAllCommandsCompleted();
				}
			}
		} else if (sequence.includes('133;C')) {
			if (this.waitingForPrompt) {
				return;
			}

			if (this.commandExecuting) {
				return;
			}

			if (this.commandQueue.length > 0) {
				setTimeout(this.processNextCommand, 500); // Small delay to ensure terminal is ready
			}
		}
	}

	private totalCommands: number = 0;

	/**
	 * Get current execution status
	 */
	public getStatus(): { executing: boolean; remaining: number; total: number } {
		return {
			executing: this.commandExecuting,
			remaining: this.commandQueue.length,
			total: this.totalCommands
		};
	}

	private checkAllCommandsCompleted(): void {
		if (this.totalCommands === 0) {
			return;
		}

		if (this.commandQueue.length === 0 && !this.commandExecuting) {
			this.options.term.writeln('\x1b[1;32mFinished executing installation commands.\x1b[0m');
			this.options.term.writeln(
				'\x1b[1;36mYou can now continue with manual commands if needed.\x1b[0m'
			);

			if (this.options.onAllCommandsCompleted) {
				this.options.onAllCommandsCompleted();
			}

			// Check if this is an automated execution and should trigger exit behavior
			if (this.options.onAllCommandsCompletedAndExit) {
				const hasOriginalCommands = this.totalCommands > 0;
				if (hasOriginalCommands) {
					this.options.onAllCommandsCompletedAndExit(true);
				}
			}

			this.totalCommands = 0;
		}
	}

	/**
	 * Stop command execution and clear the queue
	 */
	public stopExecution(): void {
		this.commandQueue = [];
		this.commandExecuting = false;
		this.totalCommands = 0;
		this.options.term.writeln('\x1b[1;33mCommand execution stopped.\x1b[0m');
	}
}
