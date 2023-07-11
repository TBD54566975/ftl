import { useParams } from 'react-router-dom'
import { VerbList } from '../verbs/VerbList'
import { useContext } from 'react'
import { schemaContext } from '../../providers/schema-provider'
import { classNames } from '../../utils/react.utils'
import { statuses } from '../../utils/style.utils'

export default function ModulePage() {
  const { id } = useParams()
  const schema = useContext(schemaContext)
  const module = schema.find(module => module.schema?.name === id)?.schema

  if (module === undefined) {
    return <></>
  }

  return (
    <>
      <div className="relative flex items-center space-x-4">
        <div className="min-w-0 flex-auto">
          <div className="flex items-center gap-x-3">
            <div className={classNames(statuses['online'], 'flex-none rounded-full p-1')}>
              <div className="h-2 w-2 rounded-full bg-current" />
            </div>

            <h2 className="min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white">
              <div className="flex gap-x-2">
                <span className="truncate">{module?.name}</span>
                <span className="text-gray-400">/</span>
                <span className="whitespace-nowrap">go</span>
                <span className="absolute inset-0" />
              </div>
            </h2>
          </div>
        </div>
      </div>
      <p className="truncate text-sm text-gray-500 pt-2">{module.comments}</p>
      <VerbList module={module} />
    </>
  )
}
