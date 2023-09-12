import { Route, Routes } from 'react-router-dom'
import { GraphPage } from './features/graph/GraphPage.tsx'
import { IDELayout } from './layout/IDELayout.tsx'
import { bgColor, textColor } from './utils/style.utils.ts'

export const App = () => {
  return (
    <div className={`h-screen flex flex-col min-w-[1024px] min-h-[600px] ${bgColor} ${textColor}`}>
      <Routes>
        <Route index element={<IDELayout />} />
        <Route path='graph' element={<GraphPage />} />
      </Routes>
    </div>
  )
}
