import * as vscode from 'vscode'
import { FtlTreeItem } from '../tree-item'

export const gotoPositionCommand = async (item: FtlTreeItem) => {
  if (!item.position) {
    return
  }

  const uri = vscode.Uri.file(item.position.filename)
  const position = new vscode.Position(Number(item.position.line) - 1, Number(item.position.column) - 1)
  const range = new vscode.Range(position, position)
  const document = await vscode.workspace.openTextDocument(uri)
  const editor = await vscode.window.showTextDocument(document)
  editor.revealRange(range, vscode.TextEditorRevealType.InCenter)
  editor.selection = new vscode.Selection(position, position)
}
