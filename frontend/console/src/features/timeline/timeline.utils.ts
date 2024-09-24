import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const eventBackgroundColorMap: Record<string, string> = {
  log: 'bg-gray-500',
  call: 'bg-indigo-500',
  ingress: 'bg-sky-400',
  deploymentCreated: 'bg-indigo-500',
  deploymentUpdated: 'bg-indigo-500',
  '': 'bg-gray-500',
}

export const eventBackgroundColor = (event: Event) => eventBackgroundColorMap[event.entry.case || '']

const eventTextColorMap: Record<string, string> = {
  log: 'text-gray-500',
  call: 'text-indigo-500',
  ingress: 'text-sky-400',
  deploymentCreated: 'text-indigo-500',
  deploymentUpdated: 'text-indigo-500',
  '': 'text-gray-500',
}

export const eventTextColor = (event: Event) => eventTextColorMap[event.entry.case || '']
