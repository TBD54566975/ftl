import { ArrowRightStartOnRectangleIcon } from '@heroicons/react/24/outline'
import { RightPanelAttribute } from '../../components/RightPanelAttribute'
import { Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from '../console/ExpandablePanel'
import { ingress, isHttpIngress, httpRequestPath, verbCalls } from './verb.utils'

export const verbPanels = (verb?: Verb) => {
  const panels = [] as ExpandablePanelProps[]

  console.log(verb?.schema)
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

  const calls = verbCalls(verb)?.map((c) => c.calls).flatMap((c) => c)
  if (calls && calls.length > 0) {
    panels.push({
      title: 'Calls',
      expanded: true,
      children: calls?.map((c, index) => (
        <div
          key={`verb-call-${index}-${c.module}-${c.name}`}
          className='flex items-center space-x-2 space-y-1'
        >
          <ArrowRightStartOnRectangleIcon className='h-4 w-4 text-blue-600' />
          <div className='truncate text-xs'>{`${c?.module}.${c?.name}`}</div>
        </div>

      )),
    })
  }

  return panels
}
