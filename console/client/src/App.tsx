import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { Layout } from './components/Layout'
import GraphPage from './features/graph/GraphPage'
import LogsPage from './features/log/LogsPage'
import ModulesPage from './features/modules/ModulesPage'
import { VerbModal } from './features/verbs/VerbModal.tsx'
import { RequestModal } from './features/requests/RequestModal.tsx'
export default function App() {
  const location = useLocation()
  const previousLocation = location.state?.previousLocation
  return (
    <>
      <Routes>
        <Route element={<Layout />}>
          <Route index
            element={<ModulesPage />}
          />
        </Route>
        {/* <Route path='graph'
          element={<GraphPage />}
        /> */}
      </Routes>
      {/* <Routes>
        <Route path='/requests/:key'
          element={<RequestModal />}
        />
        <Route path={':moduleId/verbs/:id'}
          element={<VerbModal />}
        />
      </Routes> */}
    </>
  )
}
