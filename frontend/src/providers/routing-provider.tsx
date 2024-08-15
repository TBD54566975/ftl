import { createBrowserRouter, createRoutesFromElements, Navigate, Route, RouterProvider } from 'react-router-dom'
import { Layout } from '../layout'
import { TimelinePage } from '../features/timeline/TimelinePage'
import { DeploymentsPage } from '../features/deployments/DeploymentsPage'
import { DeploymentPage } from '../features/deployments/DeploymentPage'
import { VerbPage } from '../features/verbs/VerbPage'
import { ConsolePage } from '../features/console/ConsolePage'
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
    </>
  )
)

export const RoutingProvider = () => {
  return <RouterProvider router={router} />
}
