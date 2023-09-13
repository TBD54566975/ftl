import { Navigate, Route, Routes } from 'react-router-dom'
import { GraphPage } from './features/graph/GraphPage.tsx'
import { ModulesPage } from './features/modules/ModulesPage.tsx'
import { TimelinePage } from './features/timeline/TimelinePage.tsx'
import { Layout } from './layout/Layout.tsx'
import { bgColor, textColor } from './utils/style.utils.ts'

export const App = () => {
  return (
    <div className={`h-screen flex flex-col min-w-[1024px] min-h-[600px] ${bgColor} ${textColor}`}>
      <Routes>
        <Route path='/' element={<Layout />}>
          <Route path='/' element={<Navigate to='events' replace />} />
          <Route path='events' element={<TimelinePage />} />
          <Route path='modules' element={<ModulesPage />} />
          <Route path='graph' element={<GraphPage />} />
        </Route>
      </Routes>
    </div>
  )
}
