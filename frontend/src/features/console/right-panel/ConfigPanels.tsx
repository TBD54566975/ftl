import { Config } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from '../ExpandablePanel'

export const configPanels = (config: Config) => {
  return [
    {
      title: config.config?.name,
      expanded: true,
      children: (
        <div className='flex justify-between items-center text-sm'>
          <span>Type</span>
          <span>
            <pre>{config.config?.type?.value?.case}</pre>
          </span>
        </div>
      ),
    },
  ] as ExpandablePanelProps[]
}
