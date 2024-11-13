import { Outlet } from 'react-router-dom'

export const Layout = () => {
  return (
    <div className='app'>
      <main><Outlet /></main>
    </div>
  )
}
