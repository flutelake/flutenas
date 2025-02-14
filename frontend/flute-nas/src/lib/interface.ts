export interface DirTreeNode {
    name: string;
    path: string;
children?: DirTreeNode[];
}

export interface Host {
    ID: string;
    HostIP: string;
    Hostname: string;
    AliasName: string;
    OS: string;
    OSVersion: string;
    Arch: string;
    Kernel: string;
}

export interface SambaUser {
    ID: number;
    Username: string;
    Password: string;
    Status: string;
    CreatedAt: Date
}