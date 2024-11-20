import { ResizablePanels } from '../../../../components/ResizablePanels'
import type { Database } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import { declIcon } from '../../module.utils'
import { Schema } from '../../schema/Schema'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { databasePanels } from './DatabaseRightPanels'

export const DatabasePanel = ({ value, schema, moduleName, declName }: { value: Database; schema: string; moduleName: string; declName: string }) => {
  const decl = value.database
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='Database' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
              <div className='-mx-3.5'>
                <Schema schema={schema} />
              </div>
            </div>
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('database', decl)} title={declName} />}
        rightPanelPanels={[...databasePanels(value), ...DeclDefaultPanels(schema, value.references)]}
        storageKeyPrefix='databasePanel'
      />
    </div>
  )
}
