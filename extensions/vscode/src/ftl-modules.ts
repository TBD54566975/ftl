import * as vscode from 'vscode'
import { moduleNewCommand } from './commands/module-new'
import { createClient } from './console-service'
import { ControllerService } from './protos/xyz/block/ftl/v1/ftl_connect'
import { ConsoleService } from './protos/xyz/block/ftl/v1/console/console_connect'
import { FtlTreeItem, eventToTreeItem } from './tree-item'

const controllerClient = createClient(ControllerService)
const consoleClient = createClient(ConsoleService)

let dataProvider: FtlModulesDataProvider | null

const ftlModules = new Map<string, FtlTreeItem>()

export const ftlModulesActivate = (context: vscode.ExtensionContext) => {
  dataProvider = new FtlModulesDataProvider()
  vscode.window.registerTreeDataProvider('ftlModulesView', dataProvider)
  context.subscriptions.push(
    vscode.commands.registerCommand('ftl.newModuleCommand', moduleNewCommand),
    vscode.commands.registerCommand('ftlModule.addNode', async (node: FtlTreeItem) => {
      vscode.window.showInformationMessage(`Add node command executed on ${node.label}`)
    }),
    vscode.commands.registerCommand('ftlModule.delete', (node: FtlTreeItem) => {
      vscode.window.showInformationMessage(`Delete command executed on ${node.label}`)
    })
  )
}

export const watchSchema = async (abortController: AbortController) => {
  for await (const event of controllerClient.pullSchema({ signal: abortController.signal })) {

    ftlModules.set(event.moduleName, eventToTreeItem(event))

    console.log('ftlModules:', ftlModules)

    dataProvider?.updateData(Array.from(ftlModules.values()))
  }
}

export class FtlModulesDataProvider implements vscode.TreeDataProvider<FtlTreeItem> {
  // eslint-disable-next-line max-len
  private _onDidChangeTreeData: vscode.EventEmitter<FtlTreeItem | undefined | void> = new vscode.EventEmitter<FtlTreeItem | undefined | void>()
  readonly onDidChangeTreeData: vscode.Event<FtlTreeItem | undefined | void> = this._onDidChangeTreeData.event

  private data: FtlTreeItem[] = [
    new FtlTreeItem('modules', new vscode.ThemeIcon('rocket'), vscode.TreeItemCollapsibleState.Expanded, [], 'ftlModules')
  ]

  refresh(): void {
    this._onDidChangeTreeData.fire()
  }

  updateData(newData: FtlTreeItem[]): void {
    this.data = [
      new FtlTreeItem('modules', new vscode.ThemeIcon('rocket'), vscode.TreeItemCollapsibleState.Expanded, newData, 'ftlModules')
    ]

    this.refresh()
  }

  getTreeItem(element: FtlTreeItem): vscode.TreeItem {
    return element
  }

  getChildren(element?: FtlTreeItem): Thenable<FtlTreeItem[]> {
    if (element) {
      return Promise.resolve(element.children || [])
    } else {
      return Promise.resolve(this.data)
    }
  }
}
