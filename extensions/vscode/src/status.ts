import * as vscode from "vscode"

export namespace FTLStatus {
  export const starting = (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(sync~spin) FTL`
    statusBarItem.tooltip = "FTL is starting..."
  }

  export const started = (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(zap) FTL`
    statusBarItem.tooltip = "FTL is running."
  }

  export const stopped = (statusBarItem: vscode.StatusBarItem) => {
    statusBarItem.text = `$(primitive-square) FTL`
    statusBarItem.tooltip = "FTL is stopped."
  }

  export const error = (
    statusBarItem: vscode.StatusBarItem,
    message: string
  ) => {
    statusBarItem.text = `$(error) FTL`
    statusBarItem.tooltip = message
  }
}
