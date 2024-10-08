import { InboxUploadIcon } from 'hugeicons-react'
import { RightPanelAttribute } from '../../components/RightPanelAttribute'
import type { Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../console/ExpandablePanel'
import { httpRequestPath, ingress, isHttpIngress, verbCalls } from './verb.utils'

export const verbPanels = (verb?: Verb) => {
  const panels = [] as ExpandablePanelProps[]

  if (isHttpIngress(verb)) {
    const http = ingress(verb)
    const path = httpRequestPath(verb)
    panels.push({
      title: 'HTTP Ingress',
      expanded: true,
      children: (
        <>
          <RightPanelAttribute name='Method' value={http.method} />
          <RightPanelAttribute name='Path' value={path} />
        </>
      ),
    })
  }

  const calls = verbCalls(verb)?.flatMap((c) => c.calls)
  if (calls && calls.length > 0) {
    panels.push({
      title: 'Calls',
      expanded: true,
      children: calls?.map((c, index) => (
        <div key={`verb-call-${index}-${c.module}-${c.name}`} className='flex items-center space-x-2 space-y-1'>
          <InboxUploadIcon className='h-4 w-4 text-blue-600' />
          <div className='truncate text-xs'>{`${c?.module}.${c?.name}`}</div>
        </div>
      )),
    })
  }

  return panels
}
