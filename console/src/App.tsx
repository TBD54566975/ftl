import { Navigate, Route, Routes } from 'react-router-dom'
import Logs from './features/log/Logs'
import Layout from './components/Layout'
import ModulePage from './features/modules/ModulePage'
import Navigation from './components/Navigation'
import VerbPage from './features/verbs/VerbPage'
import ModulesPage from './features/modules/ModulesPage'

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
          <Route path="logs" element={<Logs />} />
        </Route>
      </Routes>
    </>
  )
}

export default App
