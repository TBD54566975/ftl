import { GetModulesResponse, Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const HtmlId = (id: string) => ` ID="${id}"`
const TEXT = (str: string) => {
  if (str === '') return ''
  str = str.replace(/]/, '&#93;')
  return '<FONT>' + str + '</FONT>'
}
const generateRow = ({
  moduleName,
  verbName = '',
  hasCalls,
}: {
  moduleName: string
  verbName?: string
  hasCalls: boolean
}): string => {
  const callsIcon = hasCalls ? TEXT('{R}') : ''
  return `   <TR>
  <TD PORT="${verbName}" ${HtmlId(`${moduleName}.${verbName}`)}>
    <TABLE CELLPADDING="0" CELLSPACING="0" BORDER="0">
      <TR>
        <TD ALIGN="LEFT">${verbName}<FONT>  </FONT></TD>
        <TD ALIGN="RIGHT">${callsIcon}</TD>
      </TR>
    </TABLE>
  </TD>
</TR>`
}

const generateModuleContent = (module: Module): { node: string; edges: string } => {
  let edges = ''
  const moduleName = module.name
  const node = `
    ${moduleName} [
    id=${moduleName}
    label=<
      <TABLE ALIGN="LEFT" BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="5" WIDTH="250">
        <TR>
          <TD ${HtmlId(moduleName)}><FONT POINT-SIZE="18">${moduleName}</FONT></TD>
        </TR>
        ${module.verbs
          .map(({ verb }) => {
            let hasCalls = false
            verb?.metadata.forEach((metadataEntry) => {
              if (metadataEntry?.value?.case === 'calls') {
                const calls = metadataEntry.value.value.calls
                if (!hasCalls) {
                  hasCalls = Boolean(calls.length)
                }
                calls.forEach((call) => {
                  if (call.module) {
                    edges += `\n${moduleName}:${verb.name}  -> ${call.module}:${call.name};`
                  }
                })
              }
            })
            return generateRow({ moduleName, verbName: verb?.name, hasCalls })
          })
          .join('\n')}
      </TABLE>
    >
    ]`
  return { edges, node }
}

export const generateDotFile = ({ modules }: GetModulesResponse): string => {
  let nodes = ''
  let allEdges = ''
  modules.reverse().forEach((module) => {
    const { node, edges } = generateModuleContent(module)
    nodes += node
    allEdges += edges
  })
  return `
  digraph erd {
    graph [
      rankdir = "LR"
    ];
    node [
      fontsize = "16"
      fontname = "helvetica"
      shape = "plaintext"
    ];
    edge [
    ];
    ranksep = 2.0
  ${nodes}
  ${allEdges}
  }`
}
