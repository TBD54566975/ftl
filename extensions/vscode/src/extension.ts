import { ExtensionContext } from "vscode"

import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node"
import * as vscode from "vscode"
import { FTLStatus } from "./status"
import { getProjectOrWorkspaceRoot } from "./config"

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

  await startClient(context)

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


async function startClient(context: ExtensionContext) {
  console.log("Starting client")
  FTLStatus.starting(statusBarItem)
  outputChannel = vscode.window.createOutputChannel("FTL", 'log')

  const ftlConfig = vscode.workspace.getConfiguration("ftl")
  const ftlPath = ftlConfig.get<string>("executablePath") ?? "ftl"
  const userFlags = ftlConfig.get<string[]>("devCommandFlags") ?? []

  const root = await getProjectOrWorkspaceRoot()
  const flags = ["--lsp", ...userFlags]
  let serverOptions: ServerOptions = {
    run: {
      command: `${ftlPath}`,
      args: ["dev", ".", ...flags],
      options: { cwd: root }
    },
    debug: {
      command: `${ftlPath}`,
      args: ["dev", ".", ...flags],
      options: { cwd: root }
    },
  }

  outputChannel.appendLine(`Running ${ftlPath} with flags: ${flags.join(" ")}`)
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
