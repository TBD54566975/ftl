import { Navigate, Route, Routes } from 'react-router-dom'
import { DeploymentPage } from './features/deployments/DeploymentPage.tsx'
import { DeploymentsPage } from './features/deployments/DeploymentsPage.tsx'
import { GraphPage } from './features/graph/GraphPage.tsx'
import { TimelinePage } from './features/timeline/TimelinePage.tsx'
import { VerbPage } from './features/verbs/VerbPage.tsx'
import { Layout } from './layout/Layout.tsx'
import { NotFoundPage } from './layout/NotFoundPage.tsx'

export const App = () => {
  return (
    <Routes>
      <Route path='/' element={<Layout />}>
        <Route path='/' element={<Navigate to='events' replace />} />
        <Route path='events' element={<TimelinePage />} />

        <Route path='deployments' element={<DeploymentsPage />} />
        <Route path='deployments/:deploymentName' element={<DeploymentPage />} />
        <Route path='deployments/:deploymentName/verbs/:verbName' element={<VerbPage />} />

        <Route path='graph' element={<GraphPage />} />
      </Route>
      <Route path='*' element={<NotFoundPage />} />
    </Routes>
  )
}
