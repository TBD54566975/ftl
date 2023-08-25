import { Data } from '../protos/xyz/block/ftl/v1/schema/schema_pb'

export function getCodeBlock(data: Data): string {
  let codeBlock = data.comments.map((comment) => `// ${comment}`).join('\n')
  if (data.comments.length > 0) codeBlock += '\n'
  codeBlock += `type ${data.name} {`
  if (data.fields.length == 0) {
    codeBlock += `}`
    return codeBlock
  }

  codeBlock += '\n'
  data.fields.forEach((field) => {
    codeBlock += field.comments.map((comment) => `  // ${comment}`).join('\n')
    if (field.comments.length > 0) codeBlock += '\n'
    codeBlock += `  ${field.name}: ${field.type?.value.case}\n`
  })

  codeBlock += `}`
  return codeBlock
}
