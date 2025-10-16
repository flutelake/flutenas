export class FileEntry {
    name: string;
    isDir: boolean;
    size: number;
    lastMod: number;
    kind: string

    constructor(name: string, isDir: boolean, size: number, lastMod: number, kind :string) {
        this.name = name;
        this.isDir = isDir;
        this.size = size;
        this.lastMod = lastMod;
        this.kind = kind
    }
}

export class FileProgress {
    file: File;
    progress: number;
    progress2: number;
    xhr?: XMLHttpRequest;
    completedAt?: number;

    constructor(file : File, progress : number) {
        this.file = file
        this.progress = progress
        this.progress2 = 0
    }

    updateProgress(pg :number) {
        if (pg < 1) {
            this.progress = pg
            this.progress2 = this.progress + 1
        } else {
            this.progress = pg
            if (this.progress + 20 > 100) {
                this.progress2 = 100
            } else {
                this.progress2 = this.progress + 20 
            }
        }
        if (pg >= 100) {
            this.completedAt = Date.now();
        }
    } 
}

export class DiskDevice {
    Name: string;
    Type: string;
    Size: string;
    Vendor: string;
    Model: string;
    Serial: string;
    WWN: string;
    MountPoint: string;
    SpecMountPoint: string;
    FsType: string;
    UUID: string;
    PartUUID: string;
    HotPlug: boolean;
    Rota: boolean;
    IsSystemDisk: boolean;
    Labels: string[];

    constructor(
        Name: string,
        Type: string,
        Size: string,
        Vendor: string,
        Model: string,
        Serial: string,
        WWN: string,
        MountPoint: string,
        SpecMountPoint: string,
        HotPlug: boolean,
        FsType: string,
        UUID: string,
        PartUUID: string,
        Rota: boolean,
        IsSystemDisk: boolean
    ) {
        this.Name = Name;
        this.Type = Type;
        this.Size = Size;
        this.Vendor = Vendor;
        this.Model = Model;
        this.Serial = Serial;
        this.WWN = WWN;
        this.MountPoint = MountPoint;
        this.SpecMountPoint = SpecMountPoint;
        this.FsType = FsType;
        this.UUID = UUID;
        this.PartUUID = PartUUID;
        this.HotPlug = HotPlug;
        this.Rota = Rota;
        this.IsSystemDisk = IsSystemDisk;
        this.Labels = [];
        if (this.Rota) {
            this.Labels.push("HDD");
        } else {
            this.Labels.push("SSD");
        }
        if (this.HotPlug) {
            this.Labels.push("HotPlug");
        }
        if (this.IsSystemDisk) {
            this.Labels.push("System");
        }
    }

    static Unmarshal(data: any): DiskDevice {
        return new DiskDevice(
            data.Name,
            data.Type,
            data.Size,
            data.Vendor,
            data.Model,
            data.Serial,
            data.WWN,
            data.MountPoint,
            data.SpecMountPoint,
            data.HotPlug,
            data.FsType,
            data.UUID,
            data.PartUUID,
            data.Rota,
            data.IsSystemDisk
        );
    }

    static UmarshalArray(data: any): DiskDevice[] {
        return data.map((device: any) => DiskDevice.Unmarshal(device));
    }
}