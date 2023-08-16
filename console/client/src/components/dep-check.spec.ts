import { beforeAll, describe } from '@jest/globals'
import fs from 'fs'
import path from 'path'
import ts from 'typescript'
import fg from 'fast-glob'
import { expect, test } from '@jest/globals'


let cases: [dir: string, file: string][]
const srcPath = path.resolve(__dirname)
beforeAll(async () => {
  const componentDirectories: string[] = []
  const files = await fs.promises.readdir(srcPath, { withFileTypes: true })
  files.forEach(file => {
    file.isDirectory() && componentDirectories.push(file.name)
  })

  cases = componentDirectories.flatMap(dir => {
    const files = fg.sync([ path.resolve(srcPath, dir, '**/*.(ts|tsx)') ], {
      ignore: [ '**/*.spec.ts', '**/*.spec.tsx', '**/*.stories.ts', '**/*.stories.tsx' ],
      dot: true,
    })
    return files.map((file): [dir: string, file: string] => ([ dir, file ]))
  })
  return
})

describe('dependency check', () => {
  test('Check for component interdependencies', async () => {
    await Promise.all(cases.map(async ([ dir, file ]) => {
      // Parse TS file
      const node = ts.createSourceFile(
        path.basename(file),
        fs.readFileSync(file, 'utf8'),
        ts.ScriptTarget.Latest
      )

      // List of local imported  modules
      const modules: { file: string, text: string, absolutePath: string }[] = []

      node.forEachChild(child => {
        if (
          ts.SyntaxKind[child.kind] === 'ImportDeclaration' || // check imports
          ts.SyntaxKind[child.kind] === 'ExportDeclaration' // check exports
        ) {
          //@ts-ignore: not sure why this is missing from ImportDeclaration and ExportDeclaration
          const text: string = child.moduleSpecifier.text
          text.startsWith('.') &&  // If it starts with a '.' it's a local module
            modules.push({
              file,
              absolutePath: path.resolve(path.dirname(file), text),
              text,
            })
        }
      })

      const pathStart = path.resolve(srcPath, dir)
      const results = modules.find(mod => !mod.absolutePath.startsWith(pathStart))
      //@ts-ignore: custom assertion
      expect(results).noInterdependencies()
    }))
  })

})
