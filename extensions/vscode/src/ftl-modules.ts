import * as vscode from 'vscode'
import { moduleNewCommand } from './commands/module-new'
import { createClient } from './console-service'
import { ControllerService } from './protos/xyz/block/ftl/v1/ftl_connect'
import { ConsoleService } from './protos/xyz/block/ftl/v1/console/console_connect'

const controllerClient = createClient(ControllerService)
const consoleClient = createClient(ConsoleService)

let dataProvider: FtlModulesDataProvider | null

const ftlModules = new Map<string, TreeItem>()

export const ftlModulesActivate = (context: vscode.ExtensionContext) => {
  dataProvider = new FtlModulesDataProvider()
  vscode.window.registerTreeDataProvider('ftlModulesView', dataProvider)
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

export const watchSchema = async (abortController: AbortController) => {
  for await (const event of controllerClient.pullSchema({ signal: abortController.signal })) {
    ftlModules.set(event.moduleName, new TreeItem(event.moduleName, new vscode.ThemeIcon('package'), vscode.TreeItemCollapsibleState.Collapsed, [
      new TreeItem('verb1', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb'),
      new TreeItem('verb2', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb'),
    ], 'ftlModule'))

    console.log('ftlModules:', ftlModules)

    dataProvider?.updateData(Array.from(ftlModules.values()))
  }
}

export class FtlModulesDataProvider implements vscode.TreeDataProvider<TreeItem> {
  // eslint-disable-next-line max-len
  private _onDidChangeTreeData: vscode.EventEmitter<TreeItem | undefined | void> = new vscode.EventEmitter<TreeItem | undefined | void>()
  readonly onDidChangeTreeData: vscode.Event<TreeItem | undefined | void> = this._onDidChangeTreeData.event

  private data: TreeItem[] = [
    new TreeItem('modules', new vscode.ThemeIcon('rocket'), vscode.TreeItemCollapsibleState.Expanded, [], 'ftlModules')
  ]

  refresh(): void {
    this._onDidChangeTreeData.fire()
  }

  updateData(newData: TreeItem[]): void {
    this.data = [
      new TreeItem('modules', new vscode.ThemeIcon('rocket'), vscode.TreeItemCollapsibleState.Expanded, newData, 'ftlModules')
    ]

    this.refresh()
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
