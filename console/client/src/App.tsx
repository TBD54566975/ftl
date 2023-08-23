import { Route, Routes } from 'react-router-dom'
import GraphPage from './features/graph/GraphPage.tsx'
import { IDELayout } from './layout/IDELayout.tsx'
import { ShellOutlet } from './layout/ShellOutlet.tsx'

export default function App() {
  return (
    <Routes>
      <Route element={<ShellOutlet />}>
        <Route index
          element={<IDELayout />}
        />
        <Route path='graph'
          element={<GraphPage />}
        />
      </Route>
    </Routes>
  )
}
