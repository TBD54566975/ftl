import * as vscode from "vscode"
import * as fs from 'fs'
import * as path from 'path'
import { exec, execSync } from 'child_process'
import { promisify } from 'util'
import semver from 'semver'
import { lookpath } from "lookpath"

const execAsync = promisify(exec)

export const getProjectOrWorkspaceRoot = async (): Promise<string> => {
  const workspaceFolders = vscode.workspace.workspaceFolders
  if (!workspaceFolders) {
    vscode.window.showErrorMessage("FTL extension requires an open folder or workspace to work correctly.")
    return ""
  }

  return workspaceFolders[0].uri.fsPath
}

export const resolveFtlPath = async (workspaceRoot: string, config: vscode.WorkspaceConfiguration): Promise<string> => {
  const ftlPath = config.get<string>("executablePath")
  if (!ftlPath || ftlPath.trim() === '') {
    const path = await lookpath('ftl')
    if (path) {
      return path
    }
    vscode.window.showErrorMessage("FTL executable not found in PATH.")
    throw new Error("FTL executable not found in PATH.")
  }

  return path.isAbsolute(ftlPath) ? ftlPath : path.join(workspaceRoot || '', ftlPath)
}

export const getFTLVersion = async (ftlPath: string): Promise<string> => {
  try {
    const { stdout } = await execAsync(`${ftlPath} --version`)
    const version = stdout.trim()
    return version
  } catch (error) {
    throw new Error(`Failed to get FTL version\n${error}`)
  }
}

export const checkMinimumVersion = (version: string, minimumVersion: string): boolean => {
  // Always pass if the version is 'dev'
  if (version === 'dev') {
    return true
  }

  // Strip any pre-release suffixes for comparison
  const cleanVersion = version.split('-')[0]
  return semver.valid(cleanVersion) ? semver.gte(cleanVersion, minimumVersion) : false
}


export const isFTLRunning = async (ftlPath: string): Promise<boolean> => {
  try {
    await execAsync(`${ftlPath} ping`)
    return true
  } catch (error) {
    return false
  }
}
