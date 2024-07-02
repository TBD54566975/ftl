import * as vscode from 'vscode'
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from 'vscode-languageclient/node'
import { FTLStatus } from './status'

export class FTLClient {
  private clientName = 'ftl language server'
  private clientId = 'ftl'

  private statusBarItem: vscode.StatusBarItem
  private outputChannel: vscode.OutputChannel
  private client: LanguageClient | undefined
  private isClientStarting = false

  constructor(statusBar: vscode.StatusBarItem, output: vscode.OutputChannel) {
    this.statusBarItem = statusBar
    this.outputChannel = output
  }

  public async start(ftlPath: string, cwd: string, flags: string[], context: vscode.ExtensionContext) {
    if (this.client || this.isClientStarting) {
      this.outputChannel.appendLine('FTL client already running or starting')
      return
    }
    this.isClientStarting = true

    this.outputChannel.appendLine('FTL extension activated')

    const serverOptions: ServerOptions = {
      run: {
        command: `${ftlPath}`,
        args: ['dev', ...flags],
        options: { cwd: cwd }
      },
      debug: {
        command: `${ftlPath}`,
        args: ['dev', ...flags],
        options: { cwd: cwd }
      },
    }

    const clientOptions: LanguageClientOptions = {
      documentSelector: [
        { scheme: 'file', language: 'kotlin' },
        { scheme: 'file', language: 'go' },
      ],
      outputChannel: this.outputChannel,
    }

    this.client = new LanguageClient(
      this.clientId,
      this.clientName,
      serverOptions,
      clientOptions
    )

    const options = (this.client.isInDebugMode) ? serverOptions.debug : serverOptions.run
    this.outputChannel.appendLine(`Running ${ftlPath} ${options.args?.join(' ')}`)
    console.log(options)

    context.subscriptions.push(this.client)

    const buildStatus = this.client.onNotification('ftl/buildState', (message) => {
      console.log('Build status', message)
      const state = message.state

      if (state == 'building') {
        FTLStatus.buildRunning(this.statusBarItem)
      } else if (state == 'success') {
        FTLStatus.buildOK(this.statusBarItem)
      } else if (state == 'failure') {
        FTLStatus.buildError(this.statusBarItem, message.error)
      } else {
        FTLStatus.ftlError(this.statusBarItem, 'Unknown build status from FTL LSP server')
        this.outputChannel.appendLine(`Unknown build status from FTL LSP server: ${state}`)
      }
    })
    context.subscriptions.push(buildStatus)

    this.outputChannel.appendLine('Starting lsp client')
    try {
      await this.client.start()
      this.outputChannel.appendLine('Client started')
      console.log(`${this.clientName} started`)
      FTLStatus.buildOK(this.statusBarItem)
    } catch (error) {
      console.error(`Error starting ${this.clientName}: ${error}`)
      FTLStatus.ftlError(this.statusBarItem, `Error starting ${this.clientName}: ${error}`)
      this.outputChannel.appendLine(`Error starting ${this.clientName}: ${error}`)
    }

    this.isClientStarting = false
  }

  public async stop() {
    if (!this.client && !this.isClientStarting) {
      return
    }

    const timeout = 10000 // 10 seconds
    if (this.isClientStarting) {
      this.outputChannel.appendLine(`Waiting for client to complete startup before stopping`)
      const startWaitTime = Date.now()
      while (this.isClientStarting) {
        await new Promise(resolve => setTimeout(resolve, 100))
        if (Date.now() - startWaitTime > timeout) {
          this.outputChannel.appendLine(`Timeout waiting for client to start`)
          break
        }
      }
    }

    console.log('Stopping client')
    const serverProcess = this.client!['_serverProcess']

    try {
      await this.client!.stop()
      await this.client!.dispose()
      this.client = undefined
      console.log('Client stopped')
    } catch (error) {
      console.error('Error stopping client', error)
    }

    console.log('Stopping server process')
    if (serverProcess && !serverProcess.killed) {
      try {
        process.kill(serverProcess.pid, 'SIGTERM')
        // Wait a bit to see if the process terminates
        await new Promise(resolve => setTimeout(resolve, 1000))

        if (!serverProcess.killed) {
          console.log('Server process did not terminate with SIGTERM, trying SIGKILL')
          process.kill(serverProcess.pid, 'SIGKILL')
          console.log('Server process terminated with SIGKILL')
        }
      } catch (error) {
        console.log('SIGTERM failed, trying SIGKILL', error)
        try {
          // Forcefully terminate if SIGTERM fails
          process.kill(serverProcess.pid, 'SIGKILL')
          console.log('Server process terminated with SIGKILL')
        } catch (killError) {
          console.log('Failed to kill server process', killError)
        }
      }
    } else if (serverProcess && serverProcess.killed) {
      console.log('Server process was already killed')
    }

    FTLStatus.ftlStopped(this.statusBarItem)
  }
}
