import { useNavigate } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { AttributeBadge } from '../../components'
import { List } from '../../components/List'
import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { classNames } from '../../utils'
import { deploymentTextColor } from '../deployments/deployment.utils'

export const ModulesPanel = () => {
  const modules = useModules()
  const navigate = useNavigate()

  const handleModuleClick = (module: Module) => {
    navigate(`/modules/${module.name}`)
  }

  return (
    <div className='p-2'>
      <List
        items={modules.data?.modules ?? []}
        onClick={handleModuleClick}
        renderItem={(module) => (
          <div className='flex w-full' data-module-row={module.name}>
            <div className='flex gap-x-4 items-center w-1/2'>
              <div className='whitespace-nowrap'>
                <div className='flex gap-x-2 items-center'>
                  <p>{module.name}</p>
                </div>

                <p className={classNames(deploymentTextColor(module.deploymentKey), 'text-sm leading-6')}>{module.deploymentKey}</p>
              </div>
            </div>
            <div className='flex gap-x-4 items-center w-1/2 justify-end'>
              <div className='flex flex-wrap gap-2'>
                <AttributeBadge name='language' value={module.language} />
              </div>
            </div>
          </div>
        )}
      />
    </div>
  )
}
