import { useContext } from 'react'
import { useParams } from 'react-router-dom'
import { modulesContext } from '../../providers/modules-provider'
import { classNames } from '../../utils/react.utils'
import { statuses } from '../../utils/style.utils'
import { Timeline } from '../timeline/Timeline.tsx'
import { VerbList } from '../verbs/VerbList'

export default function ModulePage() {
  const { id } = useParams()
  const modules = useContext(modulesContext)
  const module = modules.modules.find(module => module?.name === id)

  if (module === undefined) {
    return <></>
  }

  return (
    <>
      <div className='relative flex items-center space-x-4'>
        <div className='min-w-0 flex-auto'>
          <div className='flex items-center gap-x-3'>
            <div className={classNames(statuses['online'], 'flex-none rounded-full p-1')}>
              <div className='h-2 w-2 rounded-full bg-current' />
            </div>

            <h2 className='min-w-0 text-sm font-semibold leading-6 text-gray-900 dark:text-white'>
              <div className='flex gap-x-2'>
                <span className='truncate'>{module?.name}</span>
                <span className='text-gray-400'>/</span>
                <span className='whitespace-nowrap'>{module.language}</span>
                <span className='absolute inset-0' />
              </div>
            </h2>
          </div>
        </div>
      </div>
      <VerbList module={module} />
      <Timeline module={module} />
    </>
  )
}
