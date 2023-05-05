import { Data, Module, Verb } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export function allVerbs(module: Module): Verb[] {
  return module?.decls.filter(decl => decl.value.case === 'verb').map(decl => decl.value.value as Verb)
}

export function getVerb(module: Module, verbName: string): Verb {
  return module?.decls.find(
    decl => decl.value.case === 'verb' && decl.value.value.name === verbName.toLocaleLowerCase(),
  )?.value.value as Verb
}

export function getData(module?: Module): Data[] {
  return module?.decls.filter(decl => decl.value.case === 'data').map(decl => decl.value.value as Data) ?? []
}
