export interface DirTreeNode {
    name: string;
    path: string;
children?: DirTreeNode[];
}

export interface Host {
    ID: number;
    HostIP: string;
    Hostname: string;
    AliasName: string;
    OS: string;
    OSVersion: string;
    Arch: string;
    Kernel: string;
}