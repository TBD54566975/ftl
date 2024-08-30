import { Navigate, Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from 'react-router-dom'
import { ConsolePage } from '../features/console/ConsolePage'
import { DeploymentPage } from '../features/deployments/DeploymentPage'
import { DeploymentsPage } from '../features/deployments/DeploymentsPage'
import { InfrastructurePage } from '../features/infrastructure/InfrastructurePage'
import { ModulePanel } from '../features/modules/ModulePanel'
import { ModulesPage, ModulesPanel } from '../features/modules/ModulesPage'
import { DeclPanel } from '../features/modules/decls/DeclPanel'
import { TimelinePage } from '../features/timeline/TimelinePage'
import { VerbPage } from '../features/verbs/VerbPage'
import { Layout } from '../layout/Layout'
import { NotFoundPage } from '../layout/NotFoundPage'

const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route path='/' element={<Layout />}>
        <Route index element={<Navigate to='events' replace />} />
        <Route path='events' element={<TimelinePage />} />
        <Route path='modules' element={<ModulesPage body={<ModulesPanel />} />} />
        <Route path='modules/:moduleName' element={<ModulesPage body={<ModulePanel />} />} />
        <Route path='modules/:moduleName/:declCase/:declName' element={<ModulesPage body={<DeclPanel />} />} />
        <Route path='deployments' element={<DeploymentsPage />} />
        <Route path='deployments/:deploymentKey' element={<DeploymentPage />} />
        <Route path='deployments/:deploymentKey/verbs/:verbName' element={<VerbPage />} />
        <Route path='graph' element={<ConsolePage />} />
        <Route path='infrastructure/*' element={<InfrastructurePage />} />
      </Route>

      <Route path='*' element={<NotFoundPage />} />
    </>,
  ),
)

export const RoutingProvider = () => {
  return <RouterProvider router={router} />
}
