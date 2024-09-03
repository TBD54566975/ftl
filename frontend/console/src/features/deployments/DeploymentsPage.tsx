import { RocketLaunchIcon } from '@heroicons/react/24/outline'
import { useModules } from '../../api/modules/use-modules'
import { Page } from '../../layout'
import { DeploymentCard } from './DeploymentCard'

export const DeploymentsPage = () => {
  const modules = useModules()

  if (!modules.isSuccess) {
    return <Page>Loading...</Page>
  }

  return (
    <Page>
      <Page.Header icon={<RocketLaunchIcon />} title='Deployments' />
      <Page.Body className='flex'>
        {modules.isLoading && <div>Loading...</div>}
        {modules.isSuccess && (
          <div className='p-4 grid grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4 overflow-y-scroll'>
            {modules.data.modules.map((module) => (
              <DeploymentCard key={module.deploymentKey} deploymentKey={module.deploymentKey} />
            ))}
          </div>
        )}
      </Page.Body>
    </Page>
  )
}
