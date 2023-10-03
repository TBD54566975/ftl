import { GetModulesResponse, Module } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { moduleTitleCls, moduleVerbCls } from '../modules.constants'
const HtmlId = (id: string) => `HREF="remove_me_url"  ID="${id}"`

const generateRow = ({
  moduleName,
  verbName = '',
  hasCalls,
}: {
  moduleName: string
  verbName?: string
  hasCalls: boolean
}): string => {
  const callsIcon = hasCalls ? '<FONT>{R}</FONT>' : ''
  return `   <TR>
  <TD ${HtmlId(`${moduleVerbCls}::${moduleName}.${verbName}`)} PORT="${verbName}" CLASS="verb">
    <TABLE CELLPADDING="0" CELLSPACING="0" BORDER="0">
      <TR>
        <TD ALIGN="LEFT">${verbName}<FONT> </FONT></TD>
        <TD ALIGN="RIGHT">${' '.repeat(4)}${callsIcon}</TD>
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
      <TABLE ALIGN="LEFT" BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="5">
        <TR>
          <TD ${HtmlId(`${moduleTitleCls}::${moduleName}`)}><FONT POINT-SIZE="18">${' '.repeat(
            4,
          )}${moduleName}${' '.repeat(4)}</FONT></TD>
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
                    edges += `\n"${moduleName}":"${verb.name}"  -> "${call.module}":"${call.name}"[
                      id = "${moduleName}.${verb.name}=>${call.module}.${call.name}"
                    ]`
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

export const generateDot = ({ modules }: GetModulesResponse): string => {
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
