import { useParams } from 'react-router-dom'
import { DeploymentPage } from '../deployments/DeploymentPage'

export const ModulePanel = () => {
  const { moduleName } = useParams()

  return <DeploymentPage moduleName={moduleName || ''} />
}
