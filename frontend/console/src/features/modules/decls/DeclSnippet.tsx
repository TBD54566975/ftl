import type { Decl } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { DataSnippet } from './DataSnippet'
import { EnumSnippet } from './EnumSnippet'
import { TypeAliasSnippet } from './TypeAliasSnippet'

export const DeclSnippet = ({ decl }: { decl: Decl }) => {
  switch (decl.value.case) {
    case 'data':
      return <DataSnippet value={decl.value.value} />
    case 'enum':
      return <EnumSnippet value={decl.value.value} />
    case 'typeAlias':
      return <TypeAliasSnippet value={decl.value.value} />
  }
  return <div className='flex-1 py-2 px-4'>under construction: {decl.value.case}</div>
}
