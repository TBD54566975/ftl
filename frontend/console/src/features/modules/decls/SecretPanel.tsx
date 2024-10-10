import type { Secret } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'

export const SecretPanel = ({ value, schema, moduleName, declName }: { value: Secret; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  return (
    <div className='py-2 px-4'>
      <PanelHeader exported={false} comments={value.comments}>
        Secret: {moduleName}.{declName}
      </PanelHeader>
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
    </div>
  )
}
