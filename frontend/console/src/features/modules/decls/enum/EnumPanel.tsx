import type { Enum } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import { Schema } from '../../schema/Schema'
import { PanelHeader } from '../PanelHeader'
import { References } from '../References'
import { enumType } from './enum.utils'

export const EnumPanel = ({ value, schema, moduleName, declName }: { value: Enum; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  const decl = value.enum
  if (!decl) {
    return
  }
  return (
    <div className='py-2 px-4'>
      <PanelHeader title={enumType(decl)} declRef={`${moduleName}.${declName}`} exported={decl.export} comments={decl.comments} />
      <div className='-mx-3.5'>
        <Schema schema={schema} />
      </div>
      <References references={value.references} />
    </div>
  )
}
