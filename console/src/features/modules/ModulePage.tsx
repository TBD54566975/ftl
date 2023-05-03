import { useParams } from 'react-router-dom'
import { classNames } from '../../utils'
import { environments, statuses } from '../../data/Types'
import { useSchema } from '../../hooks/use-schema'
import { TypeList } from '../types/TypeList'
import { VerbList } from '../verbs/VerbList'

export default function ModulePage() {
  const { id } = useParams()
  const schema = useSchema()
  const module = schema.find(module => module.schema?.name === id)?.schema

  if (module === undefined) {
    return <></>
  }

  return (
    <>
      <div className="relative flex items-center space-x-4">
        <div className="min-w-0 flex-auto">
          <div className="flex items-center gap-x-3">
            <div
              className={classNames(
                statuses['online'],
                'flex-none rounded-full p-1'
              )}
            >
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
        <div
          className={classNames(
            environments['Staging'],
            'rounded-full flex-none py-1 px-2 text-xs font-medium ring-1 ring-inset'
          )}
        >
          Staging
        </div>
      </div>
      <h2 className="text-base font-semibold dark:text-white pt-6">Verbs</h2>
      <VerbList module={module} />
      <h2 className="text-base font-semibold dark:text-white pt-6">Types</h2>
      <TypeList module={module} />
    </>
  )
}
