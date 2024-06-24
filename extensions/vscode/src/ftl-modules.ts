import * as vscode from 'vscode'

export class FtlModulesDataProvider implements vscode.TreeDataProvider<TreeItem> {
  // eslint-disable-next-line max-len
  private _onDidChangeTreeData: vscode.EventEmitter<TreeItem | undefined | void> = new vscode.EventEmitter<TreeItem | undefined | void>()
  readonly onDidChangeTreeData: vscode.Event<TreeItem | undefined | void> = this._onDidChangeTreeData.event

  private data: TreeItem[] = [new TreeItem('Item 1'), new TreeItem('Item 2')]

  refresh(): void {
    this._onDidChangeTreeData.fire()
  }

  getTreeItem(element: TreeItem): vscode.TreeItem {
    return element
  }

  getChildren(element?: TreeItem): Thenable<TreeItem[]> {
    if (element) {
      return Promise.resolve([])
    } else {
      return Promise.resolve(this.data)
    }
  }
}

class TreeItem extends vscode.TreeItem {
  constructor(label: string) {
    super(label)
    this.contextValue = 'treeItem'
  }
}
