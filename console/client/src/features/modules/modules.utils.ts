import { VerbId } from "./modules.constants"
import { VerbRef } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export const buildVerbSchema = (verbSchema: string, dataScemas: string[]): string => {
  return dataScemas.join('\n\n') + '\n\n' + verbSchema
}
export const getVerbName = (verbId: VerbId) => {
  const [_, verbName] = verbId.split('.')
  return verbName
}
export const verbRefString = (verb: VerbRef): VerbId => {
  return `${verb.module}.${verb.name}`
}