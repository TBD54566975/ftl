import { Outlet } from 'react-router-dom'
import { Navigation } from './Navigation'
import * as styles from './Layout.module.css'

export const Layout = () => {
  return (
    <>
      <Navigation />
      <main className={styles.main}>
        <Outlet />
      </main>
    </>
   
  )
}
