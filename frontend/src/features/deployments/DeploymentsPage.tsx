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
      <Page.Body className='flex'>
        <div className='p-4 grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4 overflow-y-scroll'>
          {modules.modules.map((module) => (
            <DeploymentCard key={module.deploymentKey} deploymentKey={module.deploymentKey} />
          ))}
        </div>
      </Page.Body>
    </Page>
  )
}
