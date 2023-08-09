import { Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import Navigation from './components/Navigation'
import GraphPage from './features/graph/GraphPage'
import LogsPage from './features/log/LogsPage'
import ModulePage from './features/modules/ModulePage'
import ModulesPage from './features/modules/ModulesPage'
import VerbPage from './features/verbs/VerbPage'
import RequestPage from './features/requests/RequestPage.tsx'

export default function App() {
  return (
    <>
      <Navigation />
      <Routes>
        <Route element={<Layout />}>
          <Route path='/' element={<Navigate to='/modules' replace />} />
          <Route path='modules'>
            <Route index element={<ModulesPage />} />
            <Route path={':id'} element={<ModulePage />} />
            <Route path={':moduleId/verbs/:id'} element={<VerbPage />} />
          </Route>
          <Route path='logs' element={<LogsPage />} />
          <Route path={'requests/:key'} element={<RequestPage />} />
        </Route>
        <Route path='graph' element={<GraphPage />} />
      </Routes>
    </>
  )
}
