import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { Badge } from '../../components/Badge'
import { Card } from '../../components/Card'
import { Chip } from '../../components/Chip'
import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { deploymentTextColor } from './deployment.utils'

export const DeploymentCard = ({ deploymentKey, className }: { deploymentKey: string; className?: string }) => {
  const navigate = useNavigate()
  const modules = useModules()
  const [module, setModule] = useState<Module | undefined>()

  useEffect(() => {
    if (modules.isSuccess) {
      const module = modules.data.modules.find((module) => module.deploymentKey === deploymentKey)
      setModule(module)
    }
  }, [modules.data])

  return (
    <Card key={deploymentKey} topBarColor='bg-green-500' className={className} onClick={() => navigate(`/modules/${module?.name}`)}>
      <div className='flex flex-col'>
        <div className='flex items-center'>
          <p className={`truncate flex-grow min-w-0 pr-2 ${deploymentTextColor(deploymentKey)}`}>{deploymentKey}</p>
          <Badge className='ml-auto' name={module?.language ?? ''} />
        </div>

        <div className='mt-4 gap-1 flex flex-wrap'>
          {module?.verbs.map((verb, index) => (
            <Chip key={index} className='mr-1 mb-1' name={verb.verb?.name ?? ''} />
          ))}
        </div>
      </div>
    </Card>
  )
}
