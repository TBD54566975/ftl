import {
  ArrowRightEndOnRectangleIcon,
  ArrowRightStartOnRectangleIcon,
  ArrowTopRightOnSquareIcon,
  ArrowsRightLeftIcon,
  BoltIcon,
  CodeBracketIcon,
  Cog6ToothIcon,
  LockClosedIcon,
} from '@heroicons/react/24/outline'
import { Module } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from '../ExpandablePanel'
import { CodeBlock } from '../../../components'

import { MetadataCalls } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { NavigateFunction } from 'react-router-dom'

interface InCall {
  module: string
  verb?: string
}

export const modulePanels = (
  allModules: Module[],
  module: Module,
  navigate: NavigateFunction,
): ExpandablePanelProps[] => {
  const panels = []

  if (module.verbs && module.verbs.length > 0) {
    panels.push({
      icon: BoltIcon,
      title: 'Verbs',
      expanded: true,
      children: module.verbs?.map((v) => (
        <div key={v.verb?.name} className='flex justify-between items-center'>
          <span className='text-sm truncate'>{v.verb?.name}</span>
          <button
            className='flex items-center space-x-2 hover:text-indigo-400 py-1 px-2 rounded'
            onClick={() => navigate(`/deployments/${module.deploymentKey}/verbs/${v.verb?.name}`)}
          >
            <ArrowTopRightOnSquareIcon className='size-4' />
          </button>
        </div>
      )),
    })
  }

  if (module.secrets && module.secrets.length > 0) {
    panels.push({
      icon: LockClosedIcon,
      title: 'Secrets',
      expanded: false,
      children: module.secrets.map((s, index) => (
        <div key={`secret-${s.secret?.name}-${index}`} className='flex justify-between items-center text-sm'>
          <span className='truncate pr-2'>{s.secret?.name}</span>
          <span>
            <pre>{s.secret?.type?.value?.case}</pre>
          </span>
        </div>
      )),
    })
  }

  if (module.configs && module.configs.length > 0) {
    panels.push({
      icon: Cog6ToothIcon,
      title: 'Configs',
      expanded: false,
      children: module.configs.map((c) => (
        <div key={c.config?.name} className='flex justify-between items-center text-sm'>
          <span className='truncate pr-2'>{c.config?.name}</span>
          <span>
            <pre>{c.config?.type?.value?.case}</pre>
          </span>
        </div>
      )),
    })
  }

  panels.push({
    icon: ArrowsRightLeftIcon,
    title: 'Relationships',
    expanded: false,
    children: (
      <div className='flex flex-col space-y-2'>
        {callsIn(allModules, module)?.flatMap((inCall) => (
          <div key={`in-${inCall.module}-${inCall.verb}`} className='flex items-center space-x-2'>
            <ArrowRightEndOnRectangleIcon className='h-4 w-4 text-green-600' />
            <div className='truncate text-xs'>{`${inCall?.module}.${inCall?.verb}`}</div>
          </div>
        ))}
        {callsOut(module)?.map((out, outIndex) =>
          out?.calls.map((call, callIndex) => (
            <div
              key={`out-${outIndex}-${call?.module}.${call?.name}-${callIndex}`}
              className='flex items-center space-x-2'
            >
              <ArrowRightStartOnRectangleIcon className='h-4 w-4 text-blue-600' />
              <div className='truncate text-xs'>{`${call?.module}.${call?.name}`}</div>
            </div>
          )),
        )}
      </div>
    ),
  })

  panels.push({
    icon: CodeBracketIcon,
    title: 'Schema',
    expanded: false,
    children: (
      <div className='p-0'>
        <CodeBlock code={module.schema} language='json' />
      </div>
    ),
    padding: 'p-0',
  })

  return panels
}

const callsIn = (modules: Module[], module: Module) => {
  const allCalls: InCall[] = []
  modules.forEach((m) => {
    m.verbs?.forEach((v) => {
      v.verb?.metadata
        .filter((meta) => meta.value.case === 'calls')
        .map((meta) => meta.value.value as MetadataCalls)
        .forEach((call) => {
          call.calls.forEach((c) => {
            if (c.module === module.name) {
              allCalls.push({ module: m.name, verb: v.verb?.name })
            }
          })
        })
    })
  })

  return allCalls
}

const callsOut = (module: Module) => {
  const calls = module.verbs?.flatMap((v) =>
    v.verb?.metadata.filter((meta) => meta.value.case === 'calls').map((meta) => meta.value.value as MetadataCalls),
  )

  return calls
}
