import { GetModulesResponse } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const spaces = ' '.repeat(4)
export const generateMermaidMarkdown = ({ modules }: GetModulesResponse) => {
  const subgraphs = new Map<string, Set<string>>()
  modules.forEach((module) => {
    const moduleName = module.name
    subgraphs.set(moduleName, new Set<string>())
    module.verbs.forEach((verbEntry) => {
      const verb = verbEntry.verb
      const verbName = verb?.name
      const id = `${moduleName}.${verb?.name}`
      const calls: string[] = []
      const source = `${spaces}${id}[${verbName}]`
      verb?.metadata.forEach((metadataEntry) => {
        if (metadataEntry.value.case === 'calls') {
          metadataEntry.value.value.calls.forEach((call) => {
            calls.push(`${source}-->${call.module}.${call.name}`)
          })
        }
      })
      const set = subgraphs.get(moduleName)
      calls.length ? set?.add(calls.join('\n')) : set?.add(source)
    })
  })
  const arr = ['flowchart LR']
  for (const [name, verbs] of subgraphs) {
    arr.push(`${spaces}subgraph ${name}`)
    arr.push(...verbs)
    arr.push(spaces + 'end')
  }
  return arr.join('\n')
}
