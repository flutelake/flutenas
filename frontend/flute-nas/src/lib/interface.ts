export interface DirTreeNode {
    name: string;
    path: string;
    children: DirTreeNode[];
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
    DistroID?: string;
    DistroVersion?: string;
    DistroIDLike?: string;
    PackageManager?: string;
    NFSInstalled?: boolean;
    NFSVersion?: string;
    NFSServiceStatus?: string;
    LastChecked?: Date;
}

export interface HostSystemInfo {
    HostIP: string;
    Hostname: string;
    DistroID: string;
    DistroVersion: string;
    DistroIDLike: string;
    PackageManager: string;
    Kernel: string;
    Arch: string;
    NFSInstalled: boolean;
    NFSVersion: string;
    NFSServiceStatus: string;
    InstallCommands: string[];
}

export interface NFSStatus {
    Status: string;
    Uptime?: string;
    ProcessID?: string;
}

export interface NFSStartResponse {
    Status: string;
    SetupCommands?: string[];
    InstallCommands?: string[];
    InstallRequired?: boolean;
}

export interface SambaUser {
    ID: number;
    Username: string;
    Password: string;
    Status: string;
    CreatedAt: Date;
    Permission: string;
}

export interface UserPermission {
    Username: string;
    Permission: string;
  }

export interface SambaShare {
    ID: number;
    HostIP: string;
    Name: string;
    Path: string;
    Pseudo: string | null;
    Users: UserPermission[];
    Status: string;
    CreatedAt: Date;  // ISO 日期字符串
    UpdatedAt: Date;  // ISO 日期字符串
}

export interface NFSExport {
    ID: number;
    HostIP: string;
    Name: string;
    Path: string;
    Pseudo: string;
    DefaultACL: string;
    Status: string;
    Acls: NFSExportAcl[];
    Protocols: string;
    CreatedAt: Date;  // ISO 日期字符串
    UpdatedAt: Date;  // ISO 日期字符串
}

export interface NFSExportAcl {
    IPRange: string;
    Permission: string;
}