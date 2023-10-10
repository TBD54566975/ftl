import { RocketLaunchIcon } from '@heroicons/react/24/outline'
import { useContext } from 'react'
import { Page } from '../../layout'
import { modulesContext } from '../../providers/modules-provider'
import { DeploymentCard } from './DeploymentCard'

export const DeploymentsPage = () => {
  const modules = useContext(modulesContext)

  return (
    <Page>
      <Page.Header icon={<RocketLaunchIcon />} title='Deployments' />
      <Page.Body className='p-4'>
        <div className='grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4'>
          {modules.modules.map((module) => (
            <DeploymentCard key={module.deploymentName} name={module.deploymentName} />
          ))}
        </div>
      </Page.Body>
    </Page>
  )
}
