import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const eventBackgroundColorMap: Record<string, string> = {
  log: 'bg-gray-500',
  call: 'bg-indigo-500',
  ingress: 'bg-sky-400',
  deploymentCreated: 'bg-green-500 dark:bg-green-300',
  deploymentUpdated: 'bg-green-500 dark:bg-green-300',
  '': 'bg-gray-500',
}

export const eventBackgroundColor = (event: Event) => eventBackgroundColorMap[event.entry.case || '']

const eventTextColorMap: Record<string, string> = {
  log: 'text-gray-500',
  call: 'text-indigo-500',
  ingress: 'text-sky-400',
  deploymentCreated: 'text-green-500 dark:text-green-300',
  deploymentUpdated: 'text-green-500 dark:text-green-300',
  '': 'text-gray-500',
}

export const eventTextColor = (event: Event) => eventTextColorMap[event.entry.case || '']
