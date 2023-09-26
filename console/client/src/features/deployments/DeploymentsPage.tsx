import { RocketLaunchIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useNavigate } from 'react-router-dom'
import { Card } from '../../components/Card'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'

export const DeploymentsPage = () => {
  const modules = React.useContext(modulesContext)
  const navigate = useNavigate()

  return (
    <>
      <PageHeader icon={<RocketLaunchIcon />} title='Deployments' />
      <div className='m-4 grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
        {modules.modules.map((module) => (
          <Card
            key={module.deploymentName}
            topBarColor='bg-green-500'
            onClick={() => navigate(`/deployments/${module.deploymentName}`)}
          >
            {module.name}
            <p className='text-xs text-gray-400'>{module.deploymentName}</p>
          </Card>
        ))}
      </div>
    </>
  )
}
