import { useParams } from 'react-router-dom'

export const ModulePanel = () => {
  const { moduleName } = useParams()

  return (
    <div className='flex-1 py-2 px-4'>
      <p>Module: {moduleName}</p>
    </div>
  )
}
