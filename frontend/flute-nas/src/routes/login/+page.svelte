<script lang="ts">
    import MetaTag from '../../components/MetaTag.svelte';
    import axios from 'axios'
    import JSEncrypt from 'jsencrypt';
    // import { onMount } from 'svelte';
    import { goto } from '$app/navigation';


	async function getPublicKey(): Promise<string> {
        try {
            const response = await axios.get('/v1/key');
            if (response.data.code != 0) {
                return ""
            }
            const base64DecodedKey = atob(response.data.data.key);
            // console.log("key: ", base64DecodedKey)
            return base64DecodedKey;        
        } catch (error) {
            console.log("Error fetching public key or submitting form", error);
            return ""
        }
    }
    
    function encryptWithPublicKey(publicKey :string, password: string): string {
        let encrypt = new JSEncrypt()
        encrypt.setPublicKey(publicKey);
		let result = encrypt.encrypt(password)
        if (result == false) {
			return ''
		} else {
			return result
		}
    }

    let formData = {
        username: '',
        password: ''
    };
    
    function handleSubmit(event :any) {
        // console.log(event)
        // 阻止浏览器默认的提交行为
        event.preventDefault();
       
        getPublicKey().then(publicKey => {
            console.log("key: " + publicKey)
            // 使用公钥加密密码
            if (publicKey && formData.password) {
                const encryptedPassword = encryptWithPublicKey(publicKey, formData.password);
                // console.log("encrypt key: ", encryptedPassword)
                formData.password = encryptedPassword;
            }
    
            // 在这里添加处理逻辑，例如表单验证或API调用
            // console.log("Form submitted", data);
    
            axios.post("/v1/login", {'username': formData.username, 'password': formData.password}).then(resp => {
                // console.log(resp)
                if (resp.data.code == 0) {
                    console.log("login success, nav to dashboard")
                    setTimeout(() => {
                        goto('/overview');
                    }, 1000); // 等待1000毫秒（1秒）
                }
            }).catch(err => {
                console.log("abc", err)
            })
            
        })
    }


	const path: string = '/login';
    const description: string = 'Sign in - flute nas console';
	const metaTitle: string = 'FluteNAS Web Console - Sign in';
    const subtitle: string = 'Sign in';
</script>

<MetaTag {path} {description} title={metaTitle} {subtitle} />

<div class="flex min-h-full flex-col justify-center px-6 py-12 lg:px-8">
	<div class="sm:mx-auto sm:w-full sm:max-w-sm">
		<img class="mx-auto h-10 w-auto" src="https://tailwindui.com/img/logos/mark.svg?color=indigo&shade=600" alt="Your Company"/>
		<h2 class="mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900">Sign in fluteNAS</h2>
	</div>

	<div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
		<form class="space-y-6" action="#" method="POST" on:submit={handleSubmit}>
		<div>
			<div class="flex items-center justify-between">
			<label for="username" class="block text-sm font-medium leading-6 text-gray-900">Username</label>
			</div>
			<div class="mt-2">
			<input id="username" name="username" type="text" bind:value={formData.username} autoComplete="username" required class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" />
			</div>
		</div>

		<div>
			<div class="flex items-center justify-between">
			<label for="password" class="block text-sm font-medium leading-6 text-gray-900">Password</label>
			</div>
			<div class="mt-2">
			<input id="password" name="password" type="password" bind:value={formData.password} autoComplete="password" required class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" />
			</div>
		</div>

		<div>
			<button type="submit" class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600">Sign in</button>
		</div>
		</form>
	</div>
</div>