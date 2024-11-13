import { InboxUploadIcon } from 'hugeicons-react'
import { Link } from 'react-router-dom'
import { RightPanelAttribute } from '../../components/RightPanelAttribute'
import type { Verb } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../console/ExpandablePanel'
import { Schema } from '../modules/schema/Schema'
import { type VerbRef, httpRequestPath, ingress, isHttpIngress, verbCalls } from './verb.utils'

const PanelRow = ({ verb }: { verb: VerbRef }) => {
  return (
    <Link className='flex items-center space-x-2 space-y-1 cursor-pointer' to={`/modules/${verb?.module}/verb/${verb?.name}`}>
      <InboxUploadIcon className='h-4 w-4 text-blue-600 mt-1' />
      <div className='truncate text-xs'>{`${verb?.module}.${verb?.name}`}</div>
    </Link>
  )
}

export const verbPanels = (verb?: Verb, callers?: VerbRef[]) => {
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
      children: calls?.map((c, index) => <PanelRow key={`verb-call-${index}-${c.module}-${c.name}`} verb={c} />),
    })
  }

  if (callers && callers.length > 0) {
    panels.push({
      title: 'Callers',
      expanded: true,
      children: callers.map((c, index) => <PanelRow key={`verb-caller-${index}-${c.module}-${c.name}`} verb={c} />),
    })
  }

  panels.push({
    title: 'Schema',
    expanded: true,
    children: <Schema schema={verb?.schema || ''} />,
  })

  return panels
}
