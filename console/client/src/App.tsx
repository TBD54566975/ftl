import {  Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import GraphPage from './features/graph/GraphPage'
import ModulesPage from './features/modules/ModulesPage'
export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index
          element={<ModulesPage />}
        />
        <Route path='graph'
          element={<GraphPage />}
        />
      </Route>
    </Routes>
  )
}
