import * as vscode from "vscode"
import * as fs from 'fs'
import * as path from 'path'

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
