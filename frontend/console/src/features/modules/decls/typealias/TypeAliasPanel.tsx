import { ResizablePanels } from '../../../../components/ResizablePanels'
import type { TypeAlias } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import { declIcon } from '../../module.utils'
import { Schema } from '../../schema/Schema'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { typeAliasPanels } from './TypeAliasRightPanels'

export const TypeAliasPanel = ({ value, schema, moduleName, declName }: { value: TypeAlias; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }

  const decl = value.typealias
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='TypeAlias' declRef={`${moduleName}.${declName}`} exported={decl.export} comments={decl.comments} />
              <div className='-mx-3.5'>
                <Schema schema={schema} />
              </div>
            </div>
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('typealias', decl)} title={declName} />}
        rightPanelPanels={typeAliasPanels(value, schema)}
        storageKeyPrefix='typeAliasPanel'
      />
    </div>
  )
}
