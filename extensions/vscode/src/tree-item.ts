import { PullSchemaResponse } from './protos/xyz/block/ftl/v1/ftl_pb'
import { Decl, Module, Position } from './protos/xyz/block/ftl/v1/schema/schema_pb'
import * as vscode from 'vscode'

export class FtlTreeItem extends vscode.TreeItem {
  public position: Position | undefined
  public moduleName: string | undefined

  constructor(
    label: string,
    icon: vscode.ThemeIcon,
    collapsibleState: vscode.TreeItemCollapsibleState,
    position?: Position | undefined,
    moduleName?: string | undefined,
    public readonly children?: FtlTreeItem[],
    public readonly contextValue: string = 'ftlModulesView',
    public readonly tappable: boolean = true) {
    super(label, collapsibleState)
    this.iconPath = icon
    this.position = position
    this.moduleName = moduleName
    this.contextValue = contextValue
    if (position && tappable) {
      this.command = {
        command: 'ftlModules.itemClicked',
        title: `Node Clicked: ${label}`,
        arguments: [this]
      }
    }
  }
}

export const eventToTreeItem = (event: PullSchemaResponse): FtlTreeItem => {
  const isBuiltIn = event.moduleName === 'builtin'

  return new FtlTreeItem(
    event.moduleName,
    new vscode.ThemeIcon('package'),
    vscode.TreeItemCollapsibleState.Collapsed,
    estimateModulePosition(event.schema),
    event.moduleName,
    schemaToTreeItems(isBuiltIn, event.moduleName, event.schema),
    isBuiltIn ? 'ftlBuiltinModule' : 'ftlModule'
  )
}

const schemaToTreeItems = (isBuiltIn: boolean, moduleName: string, module?: Module | undefined,): FtlTreeItem[] => {
  const treeItems: FtlTreeItem[] = []
  if (module === undefined) {
    return treeItems
  }

  module.decls.forEach(decl => {
    const item = declToTreeItem(decl, moduleName, isBuiltIn)
    if (item) {
      treeItems.push(item)
    }
  })

  return treeItems
}

const declToTreeItem = (decl: Decl, moduleName: string, isBuiltIn: boolean): FtlTreeItem | undefined => {
  const pos = decl.value.value?.pos
  const name = decl.value.value?.name ?? ''
  const collapsibleState = vscode.TreeItemCollapsibleState.None
  switch (decl.value.case) {
    case 'data':
      if (isBuiltIn) {
        return new FtlTreeItem(name, new vscode.ThemeIcon('symbol-struct'), collapsibleState, pos, moduleName, [], 'ftlData', false)
      }
      return new FtlTreeItem(name, new vscode.ThemeIcon('symbol-struct'), collapsibleState, pos, moduleName, [], 'ftlData')
    case 'verb':
      return new FtlTreeItem(name, new vscode.ThemeIcon('symbol-function'), collapsibleState, pos, moduleName, [], 'ftlVerb')
    case 'database':
      return new FtlTreeItem(name, new vscode.ThemeIcon('database'), collapsibleState, pos, moduleName, [], 'ftlDatabase')
    case 'enum':
      return new FtlTreeItem(name, new vscode.ThemeIcon('symbol-enum'), collapsibleState, pos, moduleName, [], 'ftlEnum')
    case 'typeAlias':
      return new FtlTreeItem(name, new vscode.ThemeIcon('symbol-class'), collapsibleState, pos, moduleName, [], 'ftlTypeAlias')
    case 'config':
      return new FtlTreeItem(name, new vscode.ThemeIcon('gear'), collapsibleState, pos, moduleName, [], 'ftlConfig')
    case 'secret':
      return new FtlTreeItem(name, new vscode.ThemeIcon('key'), collapsibleState, pos, moduleName, [], 'ftlSecret')
    case 'fsm':
      return new FtlTreeItem(name, new vscode.ThemeIcon('server-process'), collapsibleState, pos, moduleName, [], 'ftlFsm')
    case 'topic':
      return new FtlTreeItem(name, new vscode.ThemeIcon('broadcast'), collapsibleState, pos, moduleName, [], 'ftlTopic')
    case 'subscription':
      return new FtlTreeItem(name, new vscode.ThemeIcon('broadcast'), collapsibleState, pos, moduleName, [], 'ftlSubscription')
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
