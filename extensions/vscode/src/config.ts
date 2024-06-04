import * as vscode from "vscode"
import * as fs from 'fs'
import * as path from 'path'
import { exec, execSync } from 'child_process'
import { promisify } from 'util'
import semver from 'semver'

const execAsync = promisify(exec)

export const getProjectOrWorkspaceRoot = async (): Promise<string> => {
  const workspaceFolders = vscode.workspace.workspaceFolders
  if (!workspaceFolders) {
    vscode.window.showErrorMessage("FTL extension requires an open folder or workspace to work correctly.")
    return ""
  }

  // Check each folder for the 'ftl-project.toml' file
  for (const folder of workspaceFolders) {
    const workspaceRootPath = folder.uri.fsPath
    const ftlProjectPath = await findFileInWorkspace(workspaceRootPath, 'ftl-project.toml')
    if (ftlProjectPath) {
      return ftlProjectPath.replace('ftl-project.toml', '')
    }
  }

  return workspaceFolders[0].uri.fsPath
}

export const findFileInWorkspace = async (rootPath: string, fileName: string): Promise<string | null> => {
  try {
    const filesAndFolders = await fs.promises.readdir(rootPath, { withFileTypes: true })

    for (const dirent of filesAndFolders) {
      const fullPath = path.join(rootPath, dirent.name)
      if (dirent.isDirectory()) {
        const result = await findFileInWorkspace(fullPath, fileName)
        if (result) { return result }
      } else if (dirent.isFile() && dirent.name === fileName) {
        return fullPath
      }
    }
  } catch (error) {
    console.error('Failed to read directory:', rootPath)
  }
  return null
}

export const resolveFtlPath = (workspaceRoot: string, config: vscode.WorkspaceConfiguration): string => {
  const ftlPath = config.get<string>("executablePath")
  if (!ftlPath || ftlPath.trim() === '') {
    try {
      // Use `which` for Unix-based systems, `where` for Windows
      const command = process.platform === 'win32' ? 'where ftl' : 'which ftl'
      return execSync(command).toString().trim()
    } catch (error) {
      vscode.window.showErrorMessage('Error: ftl binary not found in PATH.')
      throw new Error('ftl binary not found in PATH')
    }
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
