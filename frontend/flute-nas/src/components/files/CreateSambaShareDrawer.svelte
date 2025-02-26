<script lang="ts">
  import { Drawer, Button, CloseButton, Label, Input, Popover, Select } from 'flowbite-svelte';
  import { InfoCircleSolid, TrashBinSolid } from 'flowbite-svelte-icons';
  import toast, { Toaster } from 'svelte-french-toast';
  import { sineIn } from 'svelte/easing';
  import { FluteAPI } from '$lib/api';
  import { type SambaUser } from '$lib/interface';
  // import DirectoryDropdown from './DirectoryDropdown.svelte';
  import DirectoryDialog from './DirectoryDialog.svelte';
	import { onMount } from 'svelte';
  import { CurrentHostIP } from '$lib/vars';
    
  
  export let selectedPath = '';
  export let hidden = true;
  export let selectDirModalFlag = false;
  // $:  = !open;
  let transitionParamsRight = {
    x: 320,
    duration: 200,
    easing: sineIn
    };
  let formData: {
    Name: string;
    Path: string;
    Pseudo: string;
    Users: SambaUser[];
  } = {
		Name: '',
		Path: '',
		Pseudo: '',
    Users: [],
	};

  let users :SambaUser[] = [{
    ID: 0,
    Username: "",
    Password: "",
    Status: "",
    CreatedAt: new Date(),
    Permission: "",
  }];

  let permissions = [
    { value: 'r', name: 'Read Only' },
    { value: 'rw', name: 'Read & Write' },
  ];

  let sambaUsers :any[] = [];

  onMount(() => {
    getSambaUsers();
  });
  function getSambaUsers() {
    const api = new FluteAPI();
    api.post('/v1/samba-user/list', {"HostIP": CurrentHostIP}).then((resp) => {
      sambaUsers = resp.data.Users.map((user: { ID: string; Username: string }) => ({
        value: user.Username,
        name: user.Username,
      }));
    }).catch((err: any) => {
      console.error('Failed to load Samba users:', err.message);
    });
  }

  function onClickAddUserInput() {
    users.push({
      ID: 0,
      Username: "",
      Password: "",
      Status: "",
      CreatedAt: new Date(),
      Permission: "",
    });
    users = users;
  }

  function onClickDelUserInput(idx: number) {
    if (users.length <= 1) {
      // 不允许全部删除掉
      return;
    }
    if (idx >= 0 && idx < users.length) {
      users.splice(idx, 1); // 从 idx 开始删除 1 个元素
    }
    users = users
  }

  function submit() {
    formData.Users = users;
    formData.Path = selectedPath;  // 确保使用选择的路径

    // 检查表单参数
    if (!formData.Name.trim()) {
      toast.error('The share name cannot be empty.');
      return;
    }

    if (!formData.Path.trim()) {
      toast.error('The path cannot be empty.');
      return;
    }

    if (!formData.Pseudo.trim()) {
      toast.error('Pseudo path cannot be empty.');
      return;
    }

    // 检查用户配置
    for (let i = 0; i < formData.Users.length; i++) {
      const user = formData.Users[i];
      if (!user.Username.trim()) {
        toast.error(`The username of the ${i + 1}th samba user cannot be empty.`);
        return;
      }
      if (!user.Permission) {
        toast.error(`The permission of the ${i + 1}th samba user cannot be empty.`);
        return;
      }
    }
    
    const api = new FluteAPI();
    api.post('/v1/samba-share/create', {
      ...formData,
      HostIP: $CurrentHostIP ? $CurrentHostIP : "127.0.0.1"
    }).then((resp) => {
      // 提交成功后的处理
      hidden = true;  // 关闭抽屉
      // 可以添加成功提示
    }).catch((err) => {
      toast.error('Failed to create Samba share:', err.message);
      // console.error();
      // 可以添加错误提示
    });
  }
   
  </script>
  
  <Drawer placement="right" transitionType="fly" {transitionParamsRight} bind:hidden={hidden} id="sidebar3" activateClickOutside={false} width='w-2/5'>
    <div class="flex items-center">
      <h5 id="drawer-label" class="inline-flex items-center mb-6 text-base font-semibold text-gray-500 uppercase dark:text-gray-400">
        <InfoCircleSolid class="w-5 h-5 me-2.5" />Create Samba Share
      </h5>
      <CloseButton on:click={() => (hidden = true)} class="mb-4 dark:text-white" />
    </div>
    <form action="#" class="mb-6">
      <div class="mb-6">
        <Label for="ShareName" class="block mb-2">Share Name</Label>
        <Input id="ShareName" name="ShareName" bind:value={formData.Name} required placeholder="" />
      </div>
      <div class="mb-6">
        <Label for="Pseudo" class="block mb-2">Pseudo Path</Label>
        <Input id="Pseudo" name="Pseudo" bind:value={formData.Pseudo} required placeholder="" />
      </div>
      <div class="mb-6">
        <Label for="Path" class="mb-2">Path</Label>
        <button on:click={() => selectDirModalFlag = true}><Input id="Pseudo" name="Pseudo" value={selectedPath} required placeholder="" /></button>
      </div>
      <!-- choose samba user -->
      <div class="mb-6">
      <Label for="user" class="mb-2">User</Label>
      {#each users as ui, index }
          <div class="grid grid-cols-5 content-start gap-6 ">
            <div class="col-span-2"> 
              <Select class="mb-2" items={sambaUsers} bind:value={ui.Username} placeholder="Select User"/>
            </div>
            <div class="col-span-2">
              <Select class="mb-2" items={permissions} bind:value={ui.Permission} placeholder="Choose Permission"/>
            </div>
            <div class="col-span-1">
              {#if index == 0 }
              <Button color="red" class="w-full"  on:click={() => onClickDelUserInput(index)} disabled>
                <TrashBinSolid size="sm" /> Del
              </Button>
              {:else}
              <Button color="red" class="w-full"  on:click={() => onClickDelUserInput(index)}>
                <TrashBinSolid size="sm" /> Del
              </Button>
              {/if}
            </div>
          </div>
      {/each}
      <Button on:click={() => onClickAddUserInput()} class="w-full">+ Add User</Button>
      </div>
      
      <Button on:click={() => submit()} class="w-full">Submit</Button>
    </form>
  </Drawer>

  <DirectoryDialog bind:open={selectDirModalFlag} bind:selectedPath={selectedPath} on:refresh_samba_user_list_msg={()=>console.log('1112')}></DirectoryDialog>