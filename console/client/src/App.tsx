import {Route, Routes} from 'react-router-dom'
import {Navigation} from './components/Navigation.tsx'
import GraphPage from './features/graph/GraphPage.tsx'
import {IDELayout} from './layout/IDELayout.tsx'
import {bgColor, textColor} from './utils/style.utils.ts'

export default function App() {
  return (
    <div
      className={`h-screen flex flex-col min-w-[1024px] min-h-[600px] ${bgColor} ${textColor}`}>
      <Navigation />
      <Routes>
        <Route
          index
          element={<IDELayout />}
        />
        <Route
          path='graph'
          element={<GraphPage />}
        />
      </Routes>
    </div>
  )
}
