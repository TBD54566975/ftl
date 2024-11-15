import type { Database } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'
import { References } from './References'

export const DatabasePanel = ({ value, schema, moduleName, declName }: { value: Database; schema: string; moduleName: string; declName: string }) => {
  return (
    <div className='py-2 px-4'>
      <PanelHeader title='Database' declRef={`${moduleName}.${declName}`} exported={false} comments={value.database?.comments} />
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
      <References references={value.references} />
    </div>
  )
}
