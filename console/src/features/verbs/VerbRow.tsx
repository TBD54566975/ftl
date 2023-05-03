import { MinusSmallIcon, PlusSmallIcon } from '@heroicons/react/20/solid'
import {
  MetadataCalls,
  Verb
} from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { Disclosure } from '@headlessui/react'
import { classNames } from '../../utils'

type Props = {
  verb?: Verb
}

export const VerbRow: React.FC<Props> = ({ verb }) => {
  const calls = verb?.metadata
    .filter(meta => meta.value.case === 'calls')
    .map(meta => meta.value.value as MetadataCalls)

  return (
    <Disclosure as="div" key={verb?.name} className="py-6">
      {({ open }) => (
        <>
          <dt>
            <Disclosure.Button className="flex w-full items-start justify-between text-left text-gray-900">
              <span className="text-base leading-7 text-sm font-semibold text-gray-400">
                {verb?.name}
              </span>
              <span className="ml-6 flex h-7 items-center">
                {open ? (
                  <MinusSmallIcon
                    className="h-6 w-6 text-gray-400"
                    aria-hidden="true"
                  />
                ) : (
                  <PlusSmallIcon
                    className="h-6 w-6 text-gray-400"
                    aria-hidden="true"
                  />
                )}
              </span>
            </Disclosure.Button>
          </dt>
          <Disclosure.Panel as="dd" className="mt-2 pr-12">
            <ul
              role="list"
              className="divide-y divide-black/5 dark:divide-white/5"
            >
              <li className="flex items-center space-x-4 py-2">
                <span
                  className={classNames(
                    'text-indigo-400 bg-indigo-400/10 ring-indigo-400/30',
                    'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
                  )}
                >
                  request
                </span>
                <div className="relative flex items-center space-x-4">
                  <div className="min-w-0 flex-auto text-indigo-500 dark:text-indigo-400">
                    <code className="text-sm">{verb?.request?.name}</code>
                  </div>
                </div>
              </li>
              <li className="flex items-center space-x-4 py-2">
                <span
                  className={classNames(
                    'text-indigo-400 bg-indigo-400/10 ring-indigo-400/30',
                    'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
                  )}
                >
                  response
                </span>
                <div className="relative flex items-center space-x-4">
                  <div className="min-w-0 flex-auto text-indigo-500 dark:text-indigo-400">
                    <code className="text-sm">{verb?.response?.name}</code>
                    <code className="text-sm">{verb?.runtime?.status}</code>
                  </div>
                </div>
              </li>
              {(calls?.length ?? 0) > 0 && (
                <li className="flex items-center space-x-4 py-2">
                  <span
                    className={classNames(
                      'text-green-400 bg-green-400/10 ring-green-400/30',
                      'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
                    )}
                  >
                    calls
                  </span>
                  <div className="relative flex items-center space-x-4">
                    <div className="min-w-0 flex-auto text-green-500 dark:text-green-400">
                      <code className="text-sm">
                        {calls?.map(call => call.calls.map(call => call.name))}
                      </code>
                    </div>
                  </div>
                </li>
              )}
            </ul>
          </Disclosure.Panel>
        </>
      )}
    </Disclosure>
  )
}
