import { ResizablePanels } from '../../../../components/ResizablePanels'
import type { Enum } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import { declIcon } from '../../module.utils'
import { Schema } from '../../schema/Schema'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { enumPanels } from './EnumRightPanels'

export const EnumPanel = ({ value, schema, moduleName, declName }: { value: Enum; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  const decl = value.enum
  if (!decl) {
    return
  }
  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='Enum' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
              <div className='-mx-3.5'>
                <Schema schema={schema} />
              </div>
            </div>
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('enum', decl)} title={declName} />}
        rightPanelPanels={[...enumPanels(value), ...DeclDefaultPanels(schema, value.references)]}
        storageKeyPrefix='enumPanel'
      />
    </div>
  )
}
