export interface DirTreeNode {
    name: string;
    path: string;
children?: DirTreeNode[];
}