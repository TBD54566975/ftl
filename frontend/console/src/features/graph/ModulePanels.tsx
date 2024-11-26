import { ArrowLeftRightIcon, FunctionIcon, InboxDownloadIcon, InboxUploadIcon, LinkSquare02Icon, Settings02Icon, SquareLock02Icon } from 'hugeicons-react'
import type { NavigateFunction } from 'react-router-dom'
import { RightPanelAttribute } from '../../components/RightPanelAttribute'
import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { ExpandablePanelProps } from '../graph/ExpandablePanel'
import { callsIn, callsOut } from '../modules/module.utils'
import { Schema } from '../modules/schema/Schema'

export const modulePanels = (allModules: Module[], module: Module, navigate: NavigateFunction): ExpandablePanelProps[] => {
  const panels = []

  if (module.verbs && module.verbs.length > 0) {
    panels.push({
      icon: FunctionIcon,
      title: 'Verbs',
      expanded: true,
      children: module.verbs?.map((v) => (
        <div key={v.verb?.name} className='flex justify-between items-center'>
          <span className='text-sm truncate'>{v.verb?.name}</span>
          <button
            type='button'
            className='flex items-center space-x-2 hover:text-indigo-400 py-1 px-2 rounded'
            onClick={() => navigate(`/deployments/${module.deploymentKey}/verbs/${v.verb?.name}`)}
          >
            <LinkSquare02Icon className='size-4' />
          </button>
        </div>
      )),
    })
  }

  if (module.secrets && module.secrets.length > 0) {
    panels.push({
      icon: SquareLock02Icon,
      title: 'Secrets',
      expanded: false,
      children: module.secrets.map((s, index) => (
        <RightPanelAttribute key={`secret-${s.secret?.name}-${index}`} name={s.secret?.name} value={s.secret?.type?.value?.case} />
      )),
    })
  }

  if (module.configs && module.configs.length > 0) {
    panels.push({
      icon: Settings02Icon,
      title: 'Configs',
      expanded: false,
      children: module.configs.map((c) => <RightPanelAttribute key={c.config?.name} name={c.config?.name} value={c.config?.type?.value?.case} />),
    })
  }

  panels.push({
    icon: ArrowLeftRightIcon,
    title: 'Relationships',
    expanded: false,
    children: (
      <div className='flex flex-col space-y-2'>
        {callsIn(allModules, module)?.flatMap((inCall) => (
          <div key={`in-${inCall.module}-${inCall.verb}`} className='flex items-center space-x-2'>
            <InboxDownloadIcon className='h-4 w-4 text-green-600' />
            <div className='truncate text-xs'>{`${inCall?.module}.${inCall?.verb}`}</div>
          </div>
        ))}
        {callsOut(module)?.map((out, outIndex) =>
          out?.calls.map((call, callIndex) => (
            <div key={`out-${outIndex}-${call?.module}.${call?.name}-${callIndex}`} className='flex items-center space-x-2'>
              <InboxUploadIcon className='h-4 w-4 text-blue-600' />
              <div className='truncate text-xs'>{`${call?.module}.${call?.name}`}</div>
            </div>
          )),
        )}
      </div>
    ),
  })

  panels.push({
    title: 'Schema',
    expanded: true,
    children: <Schema schema={module.schema} />,
  })

  return panels
}
