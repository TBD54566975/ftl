import { Navigate, Route, Routes } from 'react-router-dom'
import { ConsolePage } from './features/console/ConsolePage.tsx'
import { DeploymentPage } from './features/deployments/DeploymentPage.tsx'
import { DeploymentsPage } from './features/deployments/DeploymentsPage.tsx'
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
        <Route path='deployments/:deploymentKey' element={<DeploymentPage />} />
        <Route path='deployments/:deploymentKey/verbs/:verbName' element={<VerbPage />} />
        <Route path='console' element={<ConsolePage />} />
      </Route>
      <Route path='*' element={<NotFoundPage />} />
    </Routes>
  )
}
