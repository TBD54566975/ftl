import type { TypeAlias } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { PanelHeader } from './PanelHeader'
import { TypeAliasSnippet } from './TypeAliasSnippet'

export const TypeAliasPanel = ({ value, moduleName, declName }: { value: TypeAlias; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        Type Alias: {moduleName}.{declName}
      </PanelHeader>
      <TypeAliasSnippet value={value} />
    </div>
  )
}
