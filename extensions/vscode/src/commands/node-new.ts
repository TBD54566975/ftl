import * as vscode from 'vscode'
import { FtlTreeItem } from '../tree-item'

export const nodeNewCommand = async (item: FtlTreeItem) => {
  const uri = vscode.Uri.file(item.position?.filename || '')
  const document = vscode.workspace.textDocuments.find(doc => doc.uri.toString() === uri.toString())
  if (document === undefined) {
    return
  }

  if (document.isDirty) {
    vscode.window.showWarningMessage('Please save the document before adding a new node')
    return
  }

  //TODO: Add all the types here...
  // Also would be cool to have an httpingress type which would populate a valid //ftl:ingress ....
  const nodeType = await vscode.window.showQuickPick(['verb', 'enum'], {
    title: 'Which type of node would you like to add',
    placeHolder: 'Choose a node type',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (nodeType === undefined) {
    return
  }

  const snippet = snippetForNodeType(nodeType)

  if (snippet === '') {
    vscode.window.showErrorMessage(`No snippet available for node type ${nodeType}`)
    return
  }

  const editor = await vscode.window.showTextDocument(document)
  const edit = new vscode.WorkspaceEdit()
  const position = new vscode.Position(document.lineCount, 0)
  edit.insert(uri, position, `\n${snippet}`)

  await vscode.workspace.applyEdit(edit)
  await document.save()

  // Scroll to the bottom of the document
  const lastLine = document.lineCount - 1
  const lastLineRange = document.lineAt(lastLine).range
  editor.revealRange(lastLineRange, vscode.TextEditorRevealType.Default)
}

const snippetForNodeType = (nodeType: string): string => {
  //TODO: fill out for all node types.
  switch (nodeType) {
    case 'verb':
      return `type SampleRequest struct {
	Name string
}

type SampleResponse struct {
	Message string
}

//ftl:verb
func Sample(ctx context.Context, req SampleRequest) (SampleResponse, error) {
	return SampleResponse{Message: "Hello, world!"}, nil
}`
    case 'enum':
      return `put enum stuff here (maybe type and value enums?)`

    // Add more cases here for other node types
  }

  return ''
  // vscode.window.showInformationMessage(`Adding a new ${nodeType} node to ${document.uri.toString()}`)
}
