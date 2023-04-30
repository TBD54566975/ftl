import { Outlet } from 'react-router-dom'
const Layout = () => {
  return (
    <div className="min-h-full">
      <main>
        <div className="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
export default Layout
