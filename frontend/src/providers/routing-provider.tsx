import { Navigate, Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from 'react-router-dom'
import { ConsolePage } from '../features/console/ConsolePage'
import { DeploymentPage } from '../features/deployments/DeploymentPage'
import { DeploymentsPage } from '../features/deployments/DeploymentsPage'
import { TimelinePage } from '../features/timeline/TimelinePage'
import { VerbPage } from '../features/verbs/VerbPage'
import { Layout } from '../layout'
import { NotFoundPage } from '../layout/NotFoundPage'

const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route path='/' element={<Layout />}>
        <Route path='/' element={<Navigate to='events' replace />} />
        <Route path='events' element={<TimelinePage />} />

        <Route path='deployments' element={<DeploymentsPage />} />
        <Route path='deployments/:deploymentKey' element={<DeploymentPage />} />
        <Route path='deployments/:deploymentKey/verbs/:verbName' element={<VerbPage />} />
        <Route path='console' element={<ConsolePage />} />
      </Route>
      <Route path='*' element={<NotFoundPage />} />
    </>,
  ),
)

export const RoutingProvider = () => {
  return <RouterProvider router={router} />
}
