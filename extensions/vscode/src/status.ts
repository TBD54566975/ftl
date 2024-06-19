import * as vscode from 'vscode'

export const FTLStatus = {
  ftlStarting: (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(sync~spin) FTL`
    statusBarItem.tooltip = 'FTL is starting...'
  },
  ftlStopped: (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(primitive-square) FTL`
    statusBarItem.tooltip = 'FTL is stopped.'
  },
  ftlError: (statusBarItem: vscode.StatusBarItem, message: string) => {
    statusBarItem.text = `$(error) FTL`
    statusBarItem.tooltip = message
  },
  buildRunning: (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(gear~spin) FTL`
    statusBarItem.tooltip = 'FTL project building...'
  },
  buildOK: (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(zap) FTL`
    statusBarItem.tooltip = 'FTL project is successfully built.'
  },
  buildError: (statusBarItem: vscode.StatusBarItem, message: string) => {
    statusBarItem.text = `$(error) FTL`
    statusBarItem.tooltip = message
  },
}
