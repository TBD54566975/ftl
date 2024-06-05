import * as vscode from 'vscode'

export const FTLStatus = {
  starting: (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(sync~spin) FTL`
    statusBarItem.tooltip = 'FTL is starting...'
  },
  started: (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(zap) FTL`
    statusBarItem.tooltip = 'FTL is running.'
  },
  stopped: (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(primitive-square) FTL`
    statusBarItem.tooltip = 'FTL is stopped.'
  },
  error: (statusBarItem: vscode.StatusBarItem, message: string) => {
    statusBarItem.text = `$(error) FTL`
    statusBarItem.tooltip = message
  }
}
