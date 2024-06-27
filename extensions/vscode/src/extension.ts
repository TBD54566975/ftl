import { ExtensionContext } from 'vscode'

import * as vscode from 'vscode'
import { FTLStatus } from './status'
import { MIN_FTL_VERSION, checkMinimumVersion, getFTLVersion, getProjectOrWorkspaceRoot, isFTLRunning, resolveFtlPath } from './config'
import { FTLClient } from './client'
import { ftlModulesActivate } from './ftl-modules'

const extensionId = 'ftl'
let client: FTLClient
let statusBarItem: vscode.StatusBarItem
let outputChannel: vscode.OutputChannel

export const activate = async (context: ExtensionContext) => {
  outputChannel = vscode.window.createOutputChannel('FTL', 'log')
  outputChannel.appendLine('FTL extension activated')

  ftlModulesActivate(context)

  statusBarItem = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100
  )
  statusBarItem.command = 'ftl.statusItemClicked'
  statusBarItem.show()

  client = new FTLClient(statusBarItem, outputChannel)

  const restartCmd = vscode.commands.registerCommand(
    `${extensionId}.restart`,
    async () => {
      console.log('Restarting FTL client')
      await client.stop()
      console.log('FTL client stopped')
      await startClient(context)
      console.log('FTL client started')
    }
  )

  const stopCmd = vscode.commands.registerCommand(
    `${extensionId}.stop`,
    async () => client.stop()
  )

  const showLogsCommand = vscode.commands.registerCommand('ftl.showLogs', () => {
    outputChannel.show()
  })

  const showCommands = vscode.commands.registerCommand('ftl.statusItemClicked', () => {
    const ftlCommands = [
      { label: 'FTL: Restart Service', command: 'ftl.restart' },
      { label: 'FTL: Stop Service', command: 'ftl.stop' },
      { label: 'FTL: Show Logs', command: 'ftl.showLogs' },
    ]

    vscode.window.showQuickPick(ftlCommands, { placeHolder: 'Select an FTL command' }).then(selected => {
      if (selected) {
        vscode.commands.executeCommand(selected.command)
      }
    })
  })

  promptStartClient(context)

  context.subscriptions.push(
    restartCmd,
    stopCmd,
    statusBarItem,
    showCommands,
    showLogsCommand
  )
}

export const deactivate = async () => client.stop()

const FTLPreflightCheck = async (ftlPath: string) => {
  const ftlRunning = await isFTLRunning(ftlPath)
  if (ftlRunning) {
    vscode.window.showErrorMessage(
      'FTL is already running. Please stop the other instance and restart the service.'
    )
    return false
  }

  let version: string
  try {
    version = await getFTLVersion(ftlPath)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    vscode.window.showErrorMessage(`${error.message}`)
    return false
  }

  const versionOK = checkMinimumVersion(version, MIN_FTL_VERSION)
  if (!versionOK) {
    vscode.window.showErrorMessage(
      `FTL version ${version} is not supported. Please upgrade to at least ${MIN_FTL_VERSION}.`
    )
    return false
  }

  return true
}

const promptStartClient = async (context: vscode.ExtensionContext) => {
  const configuration = vscode.workspace.getConfiguration('ftl')
  outputChannel.appendLine(`FTL configuration: ${JSON.stringify(configuration)}`)
  const automaticallyStartServer = configuration.get<string>('automaticallyStartServer')

  FTLStatus.ftlStopped(statusBarItem)

  if (automaticallyStartServer === 'always') {
    outputChannel.appendLine(`FTL development server automatically started`)
    await startClient(context)
    return
  } else if (automaticallyStartServer === 'never') {
    outputChannel.appendLine(`FTL development server not started ('automaticallyStartServer' set to 'never' in settings.json)`)
    return
  }

  const options = ['Always', 'Yes', 'No', 'Never']
  vscode.window.showInformationMessage(
    'FTL project detected. Would you like to start the FTL development server?',
    ...options
  ).then(async (result) => {
    switch (result) {
      case 'Always':
        configuration.update('automaticallyStartServer', 'always', vscode.ConfigurationTarget.Global)
        await startClient(context)
        break
      case 'Yes':
        await startClient(context)
        break
      case 'No':
        outputChannel.appendLine('FTL development server disabled')
        FTLStatus.ftlStopped(statusBarItem)
        break
      case 'Never':
        outputChannel.appendLine('FTL development server set to never auto start')
        configuration.update('automaticallyStartServer', 'never', vscode.ConfigurationTarget.Global)
        FTLStatus.ftlStopped(statusBarItem)
        break
    }
  })
}

const startClient = async (context: ExtensionContext) => {
  FTLStatus.ftlStarting(statusBarItem)

  const ftlConfig = vscode.workspace.getConfiguration('ftl')
  const workspaceRootPath = await getProjectOrWorkspaceRoot()
  const resolvedFtlPath = await resolveFtlPath(workspaceRootPath, ftlConfig)

  outputChannel.appendLine(`VSCode workspace root path: ${workspaceRootPath}`)
  outputChannel.appendLine(`FTL path: ${resolvedFtlPath}`)

  const ftlOK = await FTLPreflightCheck(resolvedFtlPath)
  if (!ftlOK) {
    FTLStatus.ftlStopped(statusBarItem)
    return
  }

  const userFlags = ftlConfig.get<string[]>('devCommandFlags') ?? []

  const flags = [...userFlags, '--lsp']

  return client.start(resolvedFtlPath, workspaceRootPath, flags, context)
}
