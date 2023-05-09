import { Navigate, Route, Routes } from 'react-router-dom'
import LogsPage from './features/log/LogsPage'
import Layout from './components/Layout'
import ModulePage from './features/modules/ModulePage'
import Navigation from './components/Navigation'
import VerbPage from './features/verbs/VerbPage'
import ModulesPage from './features/modules/ModulesPage'
import GraphPage from './features/graph/GraphPage'

function App() {
  return (
    <>
      <Navigation />
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/modules" replace />} />
          <Route path="modules">
            <Route index element={<ModulesPage />} />
            <Route path={':id'} element={<ModulePage />} />
            <Route path={':moduleId/verbs/:id'} element={<VerbPage />} />
          </Route>
          <Route path="logs" element={<LogsPage />} />
        </Route>
        <Route path="graph" element={<GraphPage />} />
      </Routes>
    </>
  )
}

export default App
