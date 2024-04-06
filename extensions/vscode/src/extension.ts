import { ExtensionContext } from "vscode"

import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node"
import * as vscode from "vscode"
import { FTLStatus } from "./status"

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

  statusBarItem = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100
  )
  statusBarItem.command = "ftl.showLogs"
  statusBarItem.show()

  startClient(context)

  context.subscriptions.push(
    restartCmd,
    stopCmd,
    statusBarItem,
    showLogsCommand
  )
}

export async function deactivate() {
  await stopClient()
}

function startClient(context: ExtensionContext) {
  console.log("Starting client")
  FTLStatus.starting(statusBarItem)
  const ftlConfig = vscode.workspace.getConfiguration("ftl")
  const ftlPath = ftlConfig.get("executablePath") ?? "ftl"
  const userFlags = ftlConfig.get<string[]>("devCommandFlags") ?? []

  let workspaceRootPath =
    vscode.workspace.workspaceFolders &&
      vscode.workspace.workspaceFolders.length > 0
      ? vscode.workspace.workspaceFolders[0].uri.fsPath
      : null

  if (!workspaceRootPath) {
    vscode.window.showErrorMessage(
      "FTL extension requires an open folder to work correctly."
    )
    return
  }

  const flags = ["--lsp", ...userFlags]
  let serverOptions: ServerOptions = {
    run: {
      command: `${ftlPath}`,
      args: ["dev", `.`, ...flags],
    },
    debug: {
      command: `${ftlPath}`,
      args: ["dev", `.`, ...flags],
    },
  }

  console.log(serverOptions.debug.args)

  outputChannel = vscode.window.createOutputChannel("FTL Logs")
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
