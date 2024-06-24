import * as vscode from 'vscode'
import { moduleNewCommand } from './commands/module-new'

export const ftlModulesActivate = (context: vscode.ExtensionContext) => {
  const treeDataProvider = new FtlModulesDataProvider()
  vscode.window.registerTreeDataProvider('ftlModulesView', treeDataProvider)
  context.subscriptions.push(
    vscode.commands.registerCommand('ftl.newModuleCommand', moduleNewCommand),
    vscode.commands.registerCommand('ftlModule.addNode', async (node: TreeItem) => {
      vscode.window.showInformationMessage(`Add node command executed on ${node.label}`)
    }),
    vscode.commands.registerCommand('ftlModule.delete', (node: TreeItem) => {
      vscode.window.showInformationMessage(`Delete command executed on ${node.label}`)
    })
  )
}

export class FtlModulesDataProvider implements vscode.TreeDataProvider<TreeItem> {
  // eslint-disable-next-line max-len
  private _onDidChangeTreeData: vscode.EventEmitter<TreeItem | undefined | void> = new vscode.EventEmitter<TreeItem | undefined | void>()
  readonly onDidChangeTreeData: vscode.Event<TreeItem | undefined | void> = this._onDidChangeTreeData.event

  private data: TreeItem[] = [
    new TreeItem('modules', new vscode.ThemeIcon('rocket'), vscode.TreeItemCollapsibleState.Expanded, [
      new TreeItem('time', new vscode.ThemeIcon('package'), vscode.TreeItemCollapsibleState.Collapsed, [
        new TreeItem('verb1', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb'),
        new TreeItem('verb2', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb')
      ], 'ftlModule'),
      new TreeItem('echo', new vscode.ThemeIcon('package'), vscode.TreeItemCollapsibleState.Collapsed, [
        new TreeItem('verb1', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb'),
        new TreeItem('verb2', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb')
      ], 'ftlModule'),
    ], 'ftlModules')
  ]

  refresh(): void {
    this._onDidChangeTreeData.fire()
  }

  getTreeItem(element: TreeItem): vscode.TreeItem {
    return element
  }

  getChildren(element?: TreeItem): Thenable<TreeItem[]> {
    if (element) {
      return Promise.resolve(element.children || [])
    } else {
      return Promise.resolve(this.data)
    }
  }
}

class TreeItem extends vscode.TreeItem {
  constructor(
    label: string,
    icon: vscode.ThemeIcon,
    collapsibleState: vscode.TreeItemCollapsibleState,
    public readonly children?: TreeItem[],
    public readonly contextValue: string = 'ftlModulesView') {
    super(label, collapsibleState)
    this.iconPath = icon
    this.contextValue = contextValue
  }
}
