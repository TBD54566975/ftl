import { Outlet } from 'react-router-dom'
import { useStatus } from '../api/status/use-status'
import { Navigation } from './navigation/Navigation'

export const Layout = () => {
  const status = useStatus()
  const version = status.data?.controllers?.[0]?.version

  return (
    <div className='min-w-[800px] min-h-[600px] max-w-full max-h-full h-full flex flex-col dark:bg-gray-800 bg-white text-gray-700 dark:text-gray-200'>
      <Navigation version={version} />
      <main className='flex-1'>
        <Outlet />
      </main>
    </div>
  )
}
