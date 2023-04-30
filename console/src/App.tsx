import { Navigate, Route, Routes } from 'react-router-dom'
import Logs from './components/Logs'
import Layout from './components/Layout'
import Module from './components/Module'
import Navigation from './components/Navigation'
import ModuleList from './components/ModuleList'
import Verb from './components/Verb'

function App() {
  return (
    <>
      <Navigation />
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/modules" replace />} />
          <Route path="modules">
            <Route index element={<ModuleList />} />
            <Route path={':id'} element={<Module />} />
            <Route path={':moduleId/verbs/:id'} element={<Verb />} />
          </Route>
          <Route path="logs" element={<Logs />} />
        </Route>
      </Routes>
    </>
  )
}

export default App
