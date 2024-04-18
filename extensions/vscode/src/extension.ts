import { ExtensionContext } from "vscode"

import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node"
import * as vscode from "vscode"
import { FTLStatus } from "./status"
import { checkMinimumVersion, getFTLVersion, getProjectOrWorkspaceRoot, isFTLRunning } from "./config"
import path from "path"

export const MIN_FTL_VERSION = '0.169.0'

const clientName = "ftl languge server"
const clientId = "ftl"
let client: LanguageClient
let statusBarItem: vscode.StatusBarItem
let outputChannel: vscode.OutputChannel

export async function activate(context: ExtensionContext) {
  console.log('"ftl" extension activated')

  let restartCmd = vscode.commands.registerCommand(
    `${clientId}.restart`,
    async () => {
      await stopClient()
      startClient(context)
    }
  )

  let stopCmd = vscode.commands.registerCommand(
    `${clientId}.stop`,
    async () => {
      await stopClient()
    }
  )

  let showLogsCommand = vscode.commands.registerCommand("ftl.showLogs", () => {
    outputChannel.show()
  })

  let showCommands = vscode.commands.registerCommand('ftl.statusItemClicked', () => {
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

  statusBarItem = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100
  )
  statusBarItem.command = "ftl.statusItemClicked"
  statusBarItem.show()

  await promptStartClient(context)

  context.subscriptions.push(
    restartCmd,
    stopCmd,
    statusBarItem,
    showCommands,
    showLogsCommand
  )
}

export async function deactivate() {
  await stopClient()
}

async function FTLPreflightCheck(ftlPath: string) {
  const ftlRunning = await isFTLRunning(ftlPath)
  if (ftlRunning) {
    vscode.window.showErrorMessage(
      "FTL is already running. Please stop the other instance and restart the service."
    )
    return false
  }

  let version: string
  try {
    version = await getFTLVersion(ftlPath)
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

async function promptStartClient(context: vscode.ExtensionContext): Promise<void> {
  const configuration = vscode.workspace.getConfiguration('ftl')
  const startClientOption = configuration.get<string>('startClientOption')

  if (startClientOption === 'always') {
    await startClient(context)
    return
  } else if (startClientOption === 'never') {
    FTLStatus.disabled(statusBarItem)
    return
  }


  const options = ['Always', 'Yes', 'No', 'Never']
  const result = await vscode.window.showInformationMessage(
    'FTL project detected. Would you like to start the FTL development server?',
    ...options
  )

  switch (result) {
    case 'Always':
      configuration.update('startClientOption', 'always', vscode.ConfigurationTarget.Global)
      await startClient(context)
      break
    case 'Yes':
      await startClient(context)
      break
    case 'No':
      FTLStatus.disabled(statusBarItem)
      break
    case 'Never':
      configuration.update('startClientOption', 'never', vscode.ConfigurationTarget.Global)
      FTLStatus.disabled(statusBarItem)
      break
  }
}

async function startClient(context: ExtensionContext) {
  console.log("Starting client")

  const ftlConfig = vscode.workspace.getConfiguration("ftl")
  const ftlPath = ftlConfig.get<string>("executablePath") ?? "ftl"
  const workspaceRootPath = await getProjectOrWorkspaceRoot()
  const ftlAbsolutePath = path.isAbsolute(ftlPath) ? ftlPath : path.resolve(workspaceRootPath, ftlPath)

  const ftlOK = await FTLPreflightCheck(ftlAbsolutePath)
  if (!ftlOK) {
    FTLStatus.disabled(statusBarItem)
    return
  }

  FTLStatus.starting(statusBarItem)
  outputChannel = vscode.window.createOutputChannel("FTL", 'log')

  const userFlags = ftlConfig.get<string[]>("devCommandFlags") ?? []

  const flags = ["--lsp", ...userFlags]
  let serverOptions: ServerOptions = {
    run: {
      command: `${ftlAbsolutePath}`,
      args: ["dev", ".", ...flags],
      options: { cwd: workspaceRootPath }
    },
    debug: {
      command: `${ftlAbsolutePath}`,
      args: ["dev", ".", ...flags],
      options: { cwd: workspaceRootPath }
    },
  }

  outputChannel.appendLine(`Running ${ftlAbsolutePath} with flags: ${flags.join(" ")}`)
  console.log(serverOptions.debug.args)

  let clientOptions: LanguageClientOptions = {
    documentSelector: [
      { scheme: "file", language: "kotlin" },
      { scheme: "file", language: "go" },
    ],
    outputChannel,
  }

  client = new LanguageClient(
    clientId,
    clientName,
    serverOptions,
    clientOptions
  )

  console.log("Starting client")
  context.subscriptions.push(client)

  client.start().then(
    () => {
      FTLStatus.started(statusBarItem)
      outputChannel.show()
    },
    (error) => {
      console.log(`Error starting ${clientName}: ${error}`)
      FTLStatus.error(statusBarItem, `Error starting ${clientName}: ${error}`)
      outputChannel.appendLine(`Error starting ${clientName}: ${error}`)
      outputChannel.show()
    }
  )
}

async function stopClient() {
  if (!client) {
    return
  }
  console.log("Disposing client")

  client.diagnostics?.clear()
  if (client["_serverProcess"]) {
    process.kill(client["_serverProcess"].pid, "SIGINT")
  }

  //TODO: not sure why this isn't working well.
  // await client.stop();

  console.log("Client stopped")
  client.outputChannel.dispose()
  console.log("Output channel disposed")
  FTLStatus.stopped(statusBarItem)
}
