import { Outlet } from 'react-router-dom'
import { Navigation } from '../components/Navigation.tsx'
import { bgColor, textColor } from '../utils/style.utils.ts'

export const ShellOutlet = () => {
  return (
    <div className={`h-screen flex flex-col min-w-[1024px] min-h-[600px] ${bgColor} ${textColor}`}>
      <Navigation />
      <Outlet />
    </div>
  )
}
