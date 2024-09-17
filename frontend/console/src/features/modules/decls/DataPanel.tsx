import type { Data } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { DataSnippet } from './DataSnippet'
import { PanelHeader } from './PanelHeader'

export const DataPanel = ({ value, moduleName, declName }: { value: Data; moduleName: string; declName: string }) => {
  const maybeTypeParams = value.typeParameters.length === 0 ? '' : `<${value.typeParameters.map((p) => p.name).join(', ')}>`
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={value.export} comments={value.comments}>
        data: {moduleName}.{declName}
        {maybeTypeParams}
      </PanelHeader>
      <DataSnippet value={value} />
    </div>
  )
}
