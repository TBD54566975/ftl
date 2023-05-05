import { MinusSmallIcon, PlusSmallIcon } from '@heroicons/react/20/solid'
import { Data } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { Disclosure } from '@headlessui/react'
import { TypeFields } from './TypeFields'

type Props = {
  data?: Data
}

export const TypeRow: React.FC<Props> = ({ data }) => {
  return (
    <Disclosure as="div" key={data?.name} className="py-6">
      {({ open }) => (
        <>
          <dt>
            <Disclosure.Button className="flex w-full items-start justify-between text-left text-gray-900">
              <span className="text-base leading-7 text-sm font-semibold text-gray-400">{data?.name}</span>
              <span className="ml-6 flex h-7 items-center">
                {open ? (
                  <MinusSmallIcon className="h-6 w-6 text-gray-400" aria-hidden="true" />
                ) : (
                  <PlusSmallIcon className="h-6 w-6 text-gray-400" aria-hidden="true" />
                )}
              </span>
            </Disclosure.Button>
          </dt>
          <Disclosure.Panel as="dd" className="mt-2 pr-12">
            <TypeFields data={data} />
          </Disclosure.Panel>
        </>
      )}
    </Disclosure>
  )
}
