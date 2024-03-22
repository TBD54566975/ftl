import { Secret } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from '../ExpandablePanel'

export const secretPanels = (secret: Secret) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: (
        <>
          <div className='flex justify-between items-center text-sm'>
            <span>Type</span>
            <span>
              <pre>{secret.secret?.type?.value?.case}</pre>
            </span>
          </div>
        </>
      ),
    },
  ] as ExpandablePanelProps[]
}
