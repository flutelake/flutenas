// 全局变量
// import { type Host } from './interface';
import { writable } from 'svelte/store';

// Navbar.svelte 增加一层sessionStorage保存
// 这样刷新页面不会恢复初设值
export const CurrentHostIP = writable('');
export const CurrentHostIPKey = 'CurrentHostIP';
