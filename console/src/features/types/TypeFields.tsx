import { Data } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import React from 'react'
import { classNames } from '../../utils'

type Props = {
  data?: Data
}

export const TypeFields: React.FC<Props> = ({ data }) => {
  if (data?.fields.length === 0) return <></>
  return (
    <>
      <ul role="list" className="divide-y divide-black/5 dark:divide-white/5">
        {data?.fields?.map(field => (
          <li key={field.name} className="py-2">
            <div className="flex items-center space-x-4">
              <span
                className={classNames(
                  'text-indigo-400 bg-indigo-400/10 ring-indigo-400/30',
                  'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
                )}
              >
                {field.name}
              </span>
              <div className="relative flex items-center space-x-4">
                <div className="min-w-0 flex-auto text-indigo-500 dark:text-indigo-400">
                  <code className="text-sm">{field.type?.value.case}</code>
                </div>
              </div>
            </div>
            {field.comments.length > 0 &&
              field.comments.map(comment => (
                <div className="pt-4">
                  <span className="text-gray-300 text-sm">{comment}</span>
                </div>
              ))}
          </li>
        ))}
      </ul>
    </>
  )
}
