import type { TypeAlias } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'
import { References } from './References'

export const TypeAliasPanel = ({ value, schema, moduleName, declName }: { value: TypeAlias; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  return (
    <div className='py-2 px-4'>
      <PanelHeader title='Type Alias' declRef={`${moduleName}.${declName}`} exported={!!value.typealias?.export} comments={value.typealias?.comments} />
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
      <References references={value.references} />
    </div>
  )
}
