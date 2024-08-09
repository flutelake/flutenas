import axios, { type AxiosResponse } from 'axios';
import { goto } from '$app/navigation';
import type { FileProgress } from './model';
import { preprocess } from 'svelte/compiler';

export class FluteAPI {
    constructor() {}

    async post(api: string, data: any): Promise<any> { // 添加返回类型
        const that = this;
        console.log(1);
        return new Promise((resolve, reject) => {
            axios.post(api, data).then((resp: AxiosResponse) => { // 指定响应类型
                // console.log(resp)
                if (resp.data.code === 0) { // 使用严格相等
                    resolve(resp.data); // 成功时解析数据
                } else {
                    reject(new Error('Error code: ' + resp.data.code)); // 处理错误代码
                }
            }).catch(err => {
                if (err.response && err.response.status === 401) { // 使用严格相等
                    console.log("Unauthorized, nav to login page");
                    // that.nav("/login")
                    goto("/login")
                }
                reject(err); // 拒绝Promise以处理错误
            });
        });
    }

    uploadFile(path :string = '/', fp: FileProgress, handleProgress: any): Promise<any> { // 添加返回类型
        const formData = new FormData();
        formData.append(fp.file.name, fp.file);

        return new Promise((resolve, reject) => {

            axios.post('/v1/files/upload?FilePath='+path, formData, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                },
                onUploadProgress: handleProgress 
            }).then((resp: AxiosResponse) => {
                if (resp.data.code === 0) {
                    resolve(resp.data);
                } else {
                    reject(new Error('Error code: ' + resp.data.code));
                }
            }).catch(err => {
                reject(err);
            });
        });
    }
}