import { PullSchemaResponse } from 'protos/xyz/block/ftl/v1/ftl_pb'
import { Decl, Module } from 'protos/xyz/block/ftl/v1/schema/schema_pb'
import * as vscode from 'vscode'

export class FtlTreeItem extends vscode.TreeItem {
  constructor(
    label: string,
    icon: vscode.ThemeIcon,
    collapsibleState: vscode.TreeItemCollapsibleState,
    public readonly children?: FtlTreeItem[],
    public readonly contextValue: string = 'ftlModulesView') {
    super(label, collapsibleState)
    this.iconPath = icon
    this.contextValue = contextValue
  }
}

export const eventToTreeItem = (event: PullSchemaResponse): FtlTreeItem => {
  return new FtlTreeItem(
    event.moduleName, new vscode.ThemeIcon('package'),
    vscode.TreeItemCollapsibleState.Collapsed,
    schemaToTreeItems(event.schema),
    'ftlModule')
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
  // eslint-disable-next-line max-len
  // new FtlTreeItem('verb1', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb'),
  // eslint-disable-next-line max-len
  //   new FtlTreeItem('verb2', new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb'),
  return treeItems
}

const declToTreeItem = (decl: Decl): FtlTreeItem | undefined => {

  switch (decl.value.case) {
    case 'data':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-struct'), vscode.TreeItemCollapsibleState.None, [], 'ftlData')
    case 'verb':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-function'), vscode.TreeItemCollapsibleState.None, [], 'ftlVerb')
    case 'database':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('database'), vscode.TreeItemCollapsibleState.None, [], 'ftlDatabase')
    case 'enum':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-enum'), vscode.TreeItemCollapsibleState.None, [], 'ftlEnum')
    case 'typeAlias':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('symbol-class'), vscode.TreeItemCollapsibleState.None, [], 'ftlTypeAlias')
    case 'config':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('gear'), vscode.TreeItemCollapsibleState.None, [], 'ftlConfig')
    case 'secret':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('key'), vscode.TreeItemCollapsibleState.None, [], 'ftlSecret')
    case 'fsm':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('server-process'), vscode.TreeItemCollapsibleState.None, [], 'ftlFsm')
    case 'topic':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('broadcast'), vscode.TreeItemCollapsibleState.None, [], 'ftlTopic')
    case 'subscription':
      return new FtlTreeItem(decl.value.value.name, new vscode.ThemeIcon('broadcast'), vscode.TreeItemCollapsibleState.None, [], 'ftlSubscription')
  }

  return undefined
}
