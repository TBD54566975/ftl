import type { Event } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const eventBackgroundColorMap: Record<string, string> = {
  log: 'bg-gray-500',
  call: 'bg-indigo-500',
  ingress: 'bg-sky-400',
  deploymentCreated: 'bg-green-500 dark:bg-green-300',
  deploymentUpdated: 'bg-green-500 dark:bg-green-300',
  cronScheduled: 'bg-blue-500',
  '': 'bg-gray-500',
}

export const eventBackgroundColor = (event: Event) => {
  if (isError(event)) {
    return 'bg-red-500'
  }
  return eventBackgroundColorMap[event.entry.case || '']
}

const eventTextColorMap: Record<string, string> = {
  log: 'text-gray-500',
  call: 'text-indigo-500',
  ingress: 'text-sky-400',
  deploymentCreated: 'text-green-500 dark:text-green-300',
  deploymentUpdated: 'text-green-500 dark:text-green-300',
  cronScheduled: 'text-blue-500',
  '': 'text-gray-500',
}

export const eventTextColor = (event: Event) => {
  if (isError(event)) {
    return 'text-red-500'
  }
  return eventTextColorMap[event.entry.case || '']
}

const isError = (event: Event) => {
  if (event.entry.case === 'call' && event.entry.value.error) {
    return true
  }
  if (event.entry.case === 'ingress' && event.entry.value.error) {
    return true
  }
  return false
}
