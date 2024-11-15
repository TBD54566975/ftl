import { ResizablePanels } from '../../../components/ResizablePanels'
import type { Data } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { Schema } from '../schema/Schema'
import { PanelHeader } from './PanelHeader'
import { References } from './References'

export const DataPanel = ({ value, schema, moduleName, declName }: { value: Data; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  const decl = value.data
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='Data' declRef={`${moduleName}.${declName}`} exported={decl.export} comments={decl.comments} />
              <div className='-mx-3.5'>
                <Schema schema={schema} />
              </div>
            </div>
          </div>
        }
        rightPanelHeader={undefined}
        rightPanelPanels={[
          {
            title: 'References',
            expanded: true,
            children: <References references={value.references} />,
          },
        ]}
        storageKeyPrefix='dataPanel'
      />
    </div>
  )
}
