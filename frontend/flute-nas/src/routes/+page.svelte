<script>
    import { Alert } from 'flowbite-svelte';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { FluteAPI } from '$lib/api';

    onMount(async () => {
      const api = new FluteAPI()
      // 调一下接口 看看登陆状态是否正常，如果已失效则跳转到登陆的页面
      api.post("/v1/hello", {}).then(resp => {
      }).catch(err => {
        console.log(err)
      })
    });
</script>
  
<div class="p-8">
  <Alert>
    <span class="font-medium">Some Wrong!</span>
    Click <a href="/overview">here</a> jump to index page.
  </Alert>
</div>
