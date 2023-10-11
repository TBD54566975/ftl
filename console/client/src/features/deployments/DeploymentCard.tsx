import { useContext, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge } from '../../components/Badge'
import { Card } from '../../components/Card'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { deploymentTextColor } from './deployment.utils'

export const DeploymentCard = ({ deploymentName, className }: { deploymentName: string; className?: string }) => {
  const navigate = useNavigate()
  const { modules } = useContext(modulesContext)
  const [module, setModule] = useState<Module | undefined>()

  useEffect(() => {
    if (modules) {
      const module = modules.find((module) => module.deploymentName === deploymentName)
      setModule(module)
    }
  }, [modules])

  return (
    <Card
      key={deploymentName}
      topBarColor='bg-green-500'
      className={className}
      onClick={() => navigate(`/deployments/${deploymentName}`)}
    >
      <div className='flex flex-col'>
        <div className='flex items-center'>
          <p className={` ${deploymentTextColor(deploymentName)}`}>{deploymentName}</p>
          <Badge className='ml-auto' name={module?.language ?? ''} />
        </div>
        <div className='mt-2 gap-1 flex flex-wrap'>
          {module?.verbs.map((verb, index) => (
            <span key={index} className='text-xs rounded mr-1 px-1 py-0.5 border border-indigo-500'>
              {verb.verb?.name}
            </span>
          ))}
        </div>
      </div>
    </Card>
  )
}
