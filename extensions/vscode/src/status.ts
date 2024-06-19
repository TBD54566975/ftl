import * as vscode from 'vscode'

const resetColors = (statusBarItem: vscode.StatusBarItem) => {
  statusBarItem.backgroundColor = undefined
  statusBarItem.color = undefined
}

const errorColors = (statusBarItem: vscode.StatusBarItem) => {
  statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.errorBackground')
  statusBarItem.color = new vscode.ThemeColor('statusBarItem.errorForeground')
}

const loadingColors = (statusBarItem: vscode.StatusBarItem) => {
  statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground')
  statusBarItem.color = new vscode.ThemeColor('statusBarItem.warningForeground')
}

export const FTLStatus = {
  ftlStarting: (statusBarItem: vscode.StatusBarItem) => {
    loadingColors(statusBarItem)
    statusBarItem.text = `$(sync~spin) FTL`
    statusBarItem.tooltip = 'FTL is starting...'
  },
  ftlStopped: (statusBarItem: vscode.StatusBarItem) => {
    resetColors(statusBarItem)
    statusBarItem.text = `$(primitive-square) FTL`
    statusBarItem.tooltip = 'FTL is stopped.'
  },
  ftlError: (statusBarItem: vscode.StatusBarItem, message: string) => {
    errorColors(statusBarItem)
    statusBarItem.text = `$(error) FTL`
    statusBarItem.tooltip = message
  },
  buildRunning: (statusBarItem: vscode.StatusBarItem) => {
    loadingColors(statusBarItem)
    statusBarItem.text = `$(gear~spin) FTL`
    statusBarItem.tooltip = 'FTL project building...'
  },
  buildOK: (statusBarItem: vscode.StatusBarItem) => {
    resetColors(statusBarItem)
    statusBarItem.text = `$(zap) FTL`
    statusBarItem.tooltip = 'FTL project is successfully built.'
  },
  buildError: (statusBarItem: vscode.StatusBarItem, message: string) => {
    errorColors(statusBarItem)
    statusBarItem.text = `$(error) FTL`
    statusBarItem.tooltip = message
  },
}
