import { Route, Routes } from 'react-router-dom'
import Navigation from './components/Navigation.tsx'
import GraphPage from './features/graph/GraphPage.tsx'
import RequestPage from './features/requests/RequestPage.tsx'
import VerbPage from './features/verbs/VerbPage.tsx'
import { IDELayout } from './layout/IDELayout.tsx'
import { bgColor, textColor } from './utils/style.utils.ts'

export default function App() {
  return (

    <div className={`h-screen flex flex-col min-w-[1024px] min-h-[600px] ${bgColor} ${textColor}`}>
      <Navigation />
      <Routes>
        <Route element={<IDELayout />}>
          <Route path='/' />
        </Route>
        <Route path='graph'
          element={<GraphPage />}
        />
        <Route path={'requests/:key'}
          element={<RequestPage />}
        />
        <Route path={'modules/:moduleId/verbs/:id'}
          element={<VerbPage />}
        />
      </Routes>
    </div>
  )
}
