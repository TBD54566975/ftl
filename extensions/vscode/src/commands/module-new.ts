import * as vscode from 'vscode'
import { validateNotEmpty } from './utils'
import * as fs from 'fs'
import * as toml from 'toml'
import { exec } from 'child_process'
import * as util from 'util'

const execPromise = util.promisify(exec)

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

  const modulesPath = await getModulesPath()
  if (!modulesPath) {
    vscode.window.showErrorMessage('Could not find the modules directory')
    return
  }

  await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: 'Creating new FTL module',
      cancellable: true
    },
    async (progress, _) => {
      progress.report({ message: `Creating new FTL '${language}' module '${name}'` })
      try {
        const { stderr } = await execPromise(`ftl new ${language} . ${name}`, { cwd: modulesPath })
        if (stderr) {
          vscode.window.showErrorMessage(`Error: ${stderr}`)
          return
        }
      } catch (error) {
        vscode.window.showErrorMessage(`Failed to execute command: ${error}`)
        return
      }

      // fake a delay to wait for new command to complete
      progress.report({ message: `Deploying module ${name}...` })
      await new Promise<void>(resolve => setTimeout(resolve, 6000))

      vscode.window.showInformationMessage(`Module '${name}' created successfully`)
    }
  )
}

const getModulesPath = async () => {
  const workspaceFolders = vscode.workspace.workspaceFolders
  if (!workspaceFolders) {
    vscode.window.showErrorMessage('No workspace is open')
    return
  }

  const workspacePath = workspaceFolders[0].uri.fsPath
  const ftlProjectPath = `${workspacePath}/ftl-project.toml`
  if (!fs.existsSync(ftlProjectPath)) {
    return workspacePath
  }

  const fileContent = fs.readFileSync(ftlProjectPath, 'utf8')
  const parsedToml = toml.parse(fileContent)
  const moduleDirs = parsedToml['module-dirs']
  if (moduleDirs) {
    return `${workspacePath}/${moduleDirs}`
  }

  return `${workspacePath}`
}
