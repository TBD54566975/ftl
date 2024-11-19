import { ResizablePanels } from '../../../../components/ResizablePanels'
import type { Topic } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import { declIcon } from '../../module.utils'
import { Schema } from '../../schema/Schema'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { topicPanels } from './TopicRightPanels'

export const TopicPanel = ({ value, schema, moduleName, declName }: { value: Topic; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }

  const decl = value.topic
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
        rightPanelHeader={<RightPanelHeader Icon={declIcon('topic', decl)} title={declName} />}
        rightPanelPanels={[...topicPanels(value), ...DeclDefaultPanels(schema, value.references)]}
        storageKeyPrefix='topicPanel'
      />
    </div>
  )
}
