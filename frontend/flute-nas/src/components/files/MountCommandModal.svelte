<script lang="ts">
  import { onMount } from 'svelte';
  import { Modal, Button, Label, Textarea, Alert } from 'flowbite-svelte';
  import { ClipboardSolid, CheckOutline } from 'flowbite-svelte-icons';
  import { createEventDispatcher } from 'svelte';
  import type { NFSExport } from '$lib/interface';
  
  export let open = false;
  export let share: NFSExport | null = null;
  
  const dispatch = createEventDispatcher();
  
  let copiedCommand = '';
  let activeTab = 'linux';
  let serverIP = '127.0.0.1';
  let selectedNFSVersion = 'nfs4';
  
  const nfsVersions = [
    { value: 'nfs3', label: 'NFS v3' },
    { value: 'nfs4', label: 'NFS v4' },
    { value: 'nfs4.1', label: 'NFS v4.1' }
  ];
  
  // Extract server IP from current URL on component mount
  onMount(() => {
    if (typeof window !== 'undefined') {
      const hostname = window.location.hostname;
      if (hostname && hostname !== 'localhost' && hostname !== '127.0.0.1') {
        serverIP = hostname;
      } else if (window.location.host) {
        // Extract IP from host (removes port if present)
        serverIP = window.location.host.split(':')[0];
      }
    }
  });
  
  // Create a reactive dependency on selectedNFSVersion
  $: currentNFSVersion = selectedNFSVersion;
  $: currentShare = share;
  $: currentServerIP = serverIP;
  
  // Generate commands with explicit reactive dependencies
  $: mountCommands = currentShare ? generateMountCommands(currentShare, currentServerIP, currentNFSVersion) : {};
  
  // Current command being displayed
  $: currentCommand = activeTab && mountCommands[activeTab] ? mountCommands[activeTab] : '';
  
  // Debug: Log when NFS version changes
  $: if (currentNFSVersion) {
    console.log('NFS version changed to:', currentNFSVersion);
  }
  
  // Debug: Log when current command changes
  $: if (currentCommand) {
    console.log('Current command updated:', currentCommand);
  }
  
  // Debug: Log when commands change
  $: if (currentShare && currentServerIP && currentNFSVersion) {
    console.log('Mount commands updated for:', currentShare.Name, 'Version:', currentNFSVersion, 'Server:', currentServerIP);
    console.log('Generated commands:', mountCommands);
    console.log('Current command:', currentCommand);
  }
  
  function generateMountCommands(share: NFSExport, hostIP: string, nfsVersion: string) {
    const serverIP = hostIP || '127.0.0.1';
    const sharePath = share.Pseudo;
    const shareName = share.Name;
    
    // Generate version-specific mount options
    const mountOptions = getMountOptions(nfsVersion);
    const versionOptions = getVersionOptions(nfsVersion);
    
    return {
      linux: `sudo mount -t nfs ${versionOptions} ${mountOptions} ${serverIP}:${sharePath} /mnt/${shareName}`,
      linuxPersistent: `${serverIP}:${sharePath} /mnt/${shareName} nfs ${versionOptions.replace('-o ', '')},${mountOptions.replace('-o ', '')} 0 0`,
      windows: `mount \\\\${serverIP}${sharePath.replace(/\//g, '\\\\')} Z:`,
      macos: getMacOSCommand(serverIP, sharePath, shareName, nfsVersion),
      ubuntu: `sudo apt-get update && sudo apt-get install -y nfs-common\nsudo mount -t nfs ${versionOptions} ${mountOptions} ${serverIP}:${sharePath} /mnt/${shareName}`,
      centos: `sudo yum install -y nfs-utils\nsudo mount -t nfs ${versionOptions} ${mountOptions} ${serverIP}:${sharePath} /mnt/${shareName}`
    };
  }
  
  function getMacOSCommand(serverIP: string, sharePath: string, shareName: string, nfsVersion: string): string {
    const versionMap = {
      'nfs3': '-o vers=3',
      'nfs4': '-o vers=4.0',
      'nfs4.1': '-o vers=4.1'
    };
    const versionOpt = versionMap[nfsVersion] || '-o vers=4.0';
    return `sudo mount_nfs ${versionOpt} ${serverIP}:${sharePath} /Volumes/${shareName}`;
  }
  
  function getMountOptions(nfsVersion: string): string {
    switch (nfsVersion) {
      case 'nfs3':
        return '-o proto=tcp,rsize=8192,wsize=8192,timeo=14,intr';
      case 'nfs4':
        return '-o proto=tcp,rsize=8192,wsize=8192,timeo=14,intr';
      case 'nfs4.1':
        return '-o proto=tcp,rsize=8192,wsize=8192,timeo=14,intr';
      default:
        return '-o proto=tcp,rsize=8192,wsize=8192,timeo=14,intr';
    }
  }
  
  function getVersionOptions(nfsVersion: string): string {
    switch (nfsVersion) {
      case 'nfs3':
        return '-o vers=3';
      case 'nfs4':
        return '-o vers=4.0';
      case 'nfs4.1':
        return '-o vers=4.1';
      default:
        return '-o vers=4.0';
    }
  }
  
  async function copyToClipboard(command: string, label: string) {
    console.log('Attempting to copy command:', command);
    console.log('Command length:', command.length);
    console.log('Command type:', typeof command);
    
    if (!command || command.trim() === '') {
      console.error('Cannot copy empty command');
      return;
    }
    
    try {
      // Modern approach with Clipboard API
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(command);
        console.log('Successfully copied using Clipboard API');
        copiedCommand = label;
      } else {
        // Fallback for older browsers or non-secure contexts
        console.log('Using fallback copy method');
        const textArea = document.createElement('textarea');
        textArea.value = command;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        
        try {
          document.execCommand('copy');
          console.log('Successfully copied using fallback method');
          copiedCommand = label;
        } catch (err) {
          console.error('Fallback copy failed:', err);
          throw new Error('Unable to copy text');
        } finally {
          document.body.removeChild(textArea);
        }
      }
      
      // Reset copy status after 2 seconds
      setTimeout(() => {
        copiedCommand = '';
      }, 2000);
      
    } catch (err) {
      console.error('Failed to copy command:', err);
      alert('Failed to copy command to clipboard. Please select and copy the text manually.');
    }
  }
  
  function closeModal() {
    open = false;
    copiedCommand = '';
    activeTab = 'linux';
  }
  
  function getCommandForTab(tab: string): string {
    return mountCommands[tab] || '';
  }
  
  function getTabLabel(tab: string): string {
    const labels = {
      linux: 'Linux (通用)',
      linuxPersistent: 'Linux (永久挂载)',
      windows: 'Windows',
      macos: 'macOS',
      ubuntu: 'Ubuntu',
      centos: 'CentOS/RHEL'
    };
    return labels[tab] || tab;
  }
</script>

<Modal bind:open size="xl" on:close={closeModal}>
  <div class="w-full">
    <h3 class="mb-4 text-xl font-medium text-gray-900 dark:text-white">
      NFS Mount Command - {share?.Name || 'Unknown Share'}
    </h3>
    
    {#if share}
      <div class="mb-4 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
        <p class="text-sm text-blue-800 dark:text-blue-300">
          <strong>Server:</strong> {serverIP}<br>
          <strong>Share Path:</strong> {share.Pseudo}<br>
          <strong>Default Access:</strong> {share.DefaultACL}
        </p>
      </div>
      
      <!-- NFS Version Selector -->
      <div class="mb-4">
        <Label for="nfs-version" class="mb-2">NFS Version</Label>
        <select
          id="nfs-version"
          bind:value={selectedNFSVersion}
          class="w-full px-3 py-2 bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white"
        >
          {#each nfsVersions as version}
            <option value={version.value}>{version.label}</option>
          {/each}
        </select>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {selectedNFSVersion === 'nfs3' ? 'NFS v3: Traditional, widely supported, requires portmapper' : 
           selectedNFSVersion === 'nfs4' ? 'NFS v4: Modern, stateful, better security' : 
           'NFS v4.1: Latest, parallel NFS support, enhanced performance'}
        </p>
      </div>
      
      <!-- Tab Navigation -->
      <div class="mb-4 border-b border-gray-200 dark:border-gray-700">
        <ul class="flex flex-wrap -mb-px text-sm font-medium text-center">
          {#each Object.keys(mountCommands) as tab}
            <li class="mr-2">
              <button
                on:click={() => activeTab = tab}
                class="inline-block p-4 rounded-t-lg {activeTab === tab 
                  ? 'text-blue-600 border-b-2 border-blue-600 active dark:text-blue-500 dark:border-blue-500'
                  : 'text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
                type="button"
              >
                {getTabLabel(tab)}
              </button>
            </li>
          {/each}
        </ul>
      </div>
      
      <!-- Command Display -->
      <div class="mb-4">
        <Label class="mb-2">Mount Command</Label>
        <div class="relative">
          {#key `${activeTab}-${selectedNFSVersion}-${currentShare?.ID || ''}`}
            <Textarea
              id="mount-command"
              rows="3"
              class="font-mono text-sm"
              value={currentCommand}
              readonly
              placeholder="Loading command..."
            />
          {/key}
          <Button
            size="sm"
            class="absolute top-2 right-2"
            on:click={() => {
              console.log('Copy button clicked for tab:', activeTab);
              console.log('Current command to copy:', currentCommand);
              copyToClipboard(currentCommand, activeTab);
            }}
          >
            {#if copiedCommand === activeTab}
              <CheckOutline size="sm" class="mr-1" />
              Copied!
            {:else}
              <ClipboardSolid size="sm" class="mr-1" />
              Copy
            {/if}
          </Button>
        </div>
      </div>
      
      <!-- Prerequisites -->
      <Alert color="blue" class="mb-4">
        <strong>Prerequisites:</strong>
        <ul class="mt-2 text-sm list-disc list-inside">
          <li>Ensure NFS client is installed on your system</li>
          <li>Create the mount directory if it doesn't exist (e.g., <code>sudo mkdir -p /mnt/{share.Name}</code>)</li>
          <li>For permanent mounts, add the entry to <code>/etc/fstab</code></li>
        </ul>
      </Alert>
      
      <!-- Mount Directory Creation Command -->
      {#if activeTab !== 'windows' && activeTab !== 'linuxPersistent'}
        <div class="mb-4">
          <Label class="mb-2">Create Mount Directory</Label>
          <div class="relative">
            <Textarea
              rows="1"
              class="font-mono text-sm"
              value={activeTab === 'linux' || activeTab === 'ubuntu' || activeTab === 'centos' 
                ? `sudo mkdir -p /mnt/${share.Name}` 
                : `sudo mkdir -p /Volumes/${share.Name}`}
              readonly
            />
            <Button
              size="sm"
              class="absolute top-2 right-2"
              on:click={() => copyToClipboard(
                activeTab === 'linux' || activeTab === 'ubuntu' || activeTab === 'centos' 
                  ? `sudo mkdir -p /mnt/${share.Name}` 
                  : `sudo mkdir -p /Volumes/${share.Name}`,
                'mkdir'
              )}
            >
              {#if copiedCommand === 'mkdir'}
                <CheckOutline size="sm" class="mr-1" />
                Copied!
              {:else}
                <ClipboardSolid size="sm" class="mr-1" />
                Copy
              {/if}
            </Button>
          </div>
        </div>
      {/if}
      
    {:else}
      <p class="text-gray-500 dark:text-gray-400">No share selected.</p>
    {/if}
  </div>
  
  <div slot="footer" class="flex justify-end space-x-2">
    <Button color="alternative" on:click={closeModal}>Close</Button>
  </div>
</Modal>