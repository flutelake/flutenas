import axios, { type AxiosResponse } from 'axios';
import { goto } from '$app/navigation';
import type { FileProgress } from './model';
import { formatSpeed } from '$lib/index'

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

    // uploadFile(path :string = '/', fp: FileProgress, handleProgress: any): Promise<any> { // 添加返回类型
    //     const formData = new FormData();
    //     formData.append(fp.file.name, fp.file);

    //     return new Promise((resolve, reject) => {

    //         axios.post('/v1/files/upload?FilePath='+path, formData, {
    //             headers: {
    //                 'Content-Type': 'multipart/form-data'
    //             },
    //             onUploadProgress: handleProgress 
    //         }).then((resp: AxiosResponse) => {
    //             if (resp.data.code === 0) {
    //                 resolve(resp.data);
    //             } else {
    //                 reject(new Error('Error code: ' + resp.data.code));
    //             }
    //         }).catch(err => {
    //             reject(err);
    //         });
    //     });
    // }

    uploadFile(path: string = '/', fp: FileProgress, onProgress: any): Promise<any> {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            const formData = new FormData();
            formData.append(fp.file.name, fp.file);

            let startTime: number;
            let lastLoaded = 0;
            let lastTime = Date.now();
    
            xhr.open('POST', `/v1/files/upload?FilePath=${path}`, true);
            
            xhr.upload.onprogress = function(event :any) {
                if (!startTime) {
                    startTime = Date.now();
                }
                const currentTime = Date.now();
                const elapsedTime = (currentTime - startTime) / 1000; // 秒
                const loaded = event.loaded;
                const total = event.total;

                // 计算进度百分比
                const progress = Math.round((loaded / total) * 100);

                // 计算速度
                const timeInterval = (currentTime - lastTime) / 1000; // 秒
                const loadedInterval = loaded - lastLoaded;
                const speed = loadedInterval / timeInterval / 1024; // KB/s

                // 格式化速度
                const formattedSpeed = formatSpeed(speed) // 计算网速 (KB/s)

                // 调用回调函数
                onProgress(progress, formattedSpeed);

                // 更新上次加载的数据和时间
                lastLoaded = loaded;
                lastTime = currentTime;
            }
    
            xhr.onload = function() {
                if (xhr.status === 200) {
                    const response = JSON.parse(xhr.responseText);
                    if (response.code === 0) {
                        resolve(response);
                    } else {
                        reject(new Error('错误代码：' + response.code));
                    }
                } else {
                    reject(new Error('上传失败'));
                }
            };
    
            xhr.onerror = function() {
                reject(new Error('网络错误'));
            };
    
            xhr.send(formData);
        });
    }
}