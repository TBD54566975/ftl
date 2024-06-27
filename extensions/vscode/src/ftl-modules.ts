import * as vscode from 'vscode'
import { moduleNewCommand } from './commands/module-new'
import { createClient } from './console-service'
import { ControllerService } from './protos/xyz/block/ftl/v1/ftl_connect'
import { FtlTreeItem, eventToTreeItem } from './tree-item'
import { gotoPositionCommand } from './commands/goto-position'
import { nodeNewCommand } from './commands/node-new'
import { DeploymentChangeType } from './protos/xyz/block/ftl/v1/ftl_pb'

const controllerClient = createClient(ControllerService)

let dataProvider: FtlModulesDataProvider | null

const ftlModules = new Map<string, FtlTreeItem>()

export const ftlModulesActivate = (context: vscode.ExtensionContext) => {
  dataProvider = new FtlModulesDataProvider()
  vscode.window.registerTreeDataProvider('ftlModulesView', dataProvider)
  context.subscriptions.push(
    vscode.commands.registerCommand('ftl.newModuleCommand', moduleNewCommand),
    vscode.commands.registerCommand('ftlModule.addNode', nodeNewCommand),
    vscode.commands.registerCommand('ftlModule.delete', (node: FtlTreeItem) => {
      vscode.window.showInformationMessage(`Delete command executed on ${node.label}`)
    })
  )
  vscode.commands.registerCommand('ftlModules.itemClicked', (item: FtlTreeItem) => gotoPositionCommand(item))
}

export const watchSchema = async (abortController: AbortController) => {
  for await (const event of controllerClient.pullSchema({ signal: abortController.signal })) {
    if (event.changeType === DeploymentChangeType.DEPLOYMENT_REMOVED) {
      ftlModules.delete(event.moduleName)
      dataProvider?.updateData(Array.from(ftlModules.values()))
      continue
    }
    ftlModules.set(event.moduleName, eventToTreeItem(event))
    dataProvider?.updateData(Array.from(ftlModules.values()))
  }
}

export class FtlModulesDataProvider implements vscode.TreeDataProvider<FtlTreeItem> {
  // eslint-disable-next-line max-len
  private _onDidChangeTreeData: vscode.EventEmitter<FtlTreeItem | undefined | void> = new vscode.EventEmitter<FtlTreeItem | undefined | void>()
  readonly onDidChangeTreeData: vscode.Event<FtlTreeItem | undefined | void> = this._onDidChangeTreeData.event

  private data: FtlTreeItem[] = [
    new FtlTreeItem('modules', new vscode.ThemeIcon('rocket'), vscode.TreeItemCollapsibleState.Expanded, undefined, undefined, [], 'ftlModules')
  ]

  refresh(): void {
    this._onDidChangeTreeData.fire()
  }

  updateData(newData: FtlTreeItem[]): void {
    this.data = [
      new FtlTreeItem('modules', new vscode.ThemeIcon('rocket'), vscode.TreeItemCollapsibleState.Expanded, undefined, undefined, newData, 'ftlModules')
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
