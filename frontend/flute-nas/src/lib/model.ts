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
    } 
}