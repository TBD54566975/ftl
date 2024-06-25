import { PullSchemaResponse } from './protos/xyz/block/ftl/v1/ftl_pb'
import { Decl, Module, Position } from './protos/xyz/block/ftl/v1/schema/schema_pb'
import * as vscode from 'vscode'

export class FtlTreeItem extends vscode.TreeItem {
  public position: Position | undefined

  constructor(
    label: string,
    icon: vscode.ThemeIcon,
    collapsibleState: vscode.TreeItemCollapsibleState,
    position: Position | undefined,
    public readonly children?: FtlTreeItem[],
    public readonly contextValue: string = 'ftlModulesView') {
    super(label, collapsibleState)
    this.position = position
    this.iconPath = icon
    this.contextValue = contextValue
    if (position) {
      this.command = {
        command: 'ftlModules.itemClicked',
        title: `Node Clicked: ${label}`,
        arguments: [this]
      }
    }
  }
}

export const eventToTreeItem = (event: PullSchemaResponse): FtlTreeItem => {
  return new FtlTreeItem(
    event.moduleName, new vscode.ThemeIcon('package'),
    vscode.TreeItemCollapsibleState.Collapsed,
    estimateModulePosition(event.schema),
    schemaToTreeItems(event.schema),
    'ftlModule'
  )
}

const schemaToTreeItems = (module?: Module | undefined): FtlTreeItem[] => {
  const treeItems: FtlTreeItem[] = []
  if (module === undefined) {
    return treeItems
  }

  module.decls.forEach(decl => {
    const item = declToTreeItem(decl)
    if (item) {
      treeItems.push(item)
    }
  })

  return treeItems
}

const declToTreeItem = (decl: Decl): FtlTreeItem | undefined => {
  const pos = decl.value.value?.pos
  switch (decl.value.case) {
    case 'data':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-struct'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlData')
    case 'verb':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlVerb')
    case 'database':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('database'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlDatabase')
    case 'enum':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-enum'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlEnum')
    case 'typeAlias':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-class'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlTypeAlias')
    case 'config':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('gear'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlConfig')
    case 'secret':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('key'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlSecret')
    case 'fsm':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('server-process'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlFsm')
    case 'topic':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('broadcast'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlTopic')
    case 'subscription':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('broadcast'), vscode.TreeItemCollapsibleState.None, pos, [], 'ftlSubscription')
  }

  return undefined
}

const estimateModulePosition = (module: Module | undefined): Position | undefined => {
  if (module?.name === 'builtin') {
    return undefined
  }

  const declPos = module?.decls.find(decl => decl.value.value?.pos !== undefined)?.value.value?.pos
  if (!declPos) {
    return undefined
  }

  const pos = new Position()
  pos.filename = declPos?.filename
  pos.line = BigInt(1)
  pos.column = BigInt(1)
  return pos
}
