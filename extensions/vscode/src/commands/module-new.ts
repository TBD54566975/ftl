import * as vscode from 'vscode'
import { validateNotEmpty } from './utils'

export const moduleNewCommand = async () => {
  const name = await vscode.window.showInputBox({
    title: 'Enter a name for your module',
    placeHolder: 'Module name',
    validateInput: validateNotEmpty,
    ignoreFocusOut: true
  })
  if (!name) {
    return
  }

  const language = await vscode.window.showQuickPick(['go', 'kotlin'], {
    title: 'Choose a language for your module',
    placeHolder: 'Choose a language',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (language === undefined) {
    return
  }

  await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: 'Processing',
      cancellable: true
    },
    async (progress, token) => {
      token.onCancellationRequested(() => {
        vscode.window.showInformationMessage('User canceled the long running operation')
      })

      progress.report({ message: `Creating new FTL ${language} module ${name}` })

      await new Promise<void>(resolve => setTimeout(resolve, 1000))
      progress.report({ message: 'Working...' })

      await new Promise<void>(resolve => setTimeout(resolve, 2000))
    }
  )
}
