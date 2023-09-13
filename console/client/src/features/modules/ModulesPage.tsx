import { ViewModuleRounded } from '@mui/icons-material'
import { PageHeader } from '../../components/PageHeader'
import { ModulesList } from './ModulesList'

export const ModulesPage = () => {
  return (
    <>
      <div className='w-full m-0'>
        <PageHeader icon={<ViewModuleRounded />} title='Modules' />
        <ModulesList />
      </div>
    </>
  )
}
