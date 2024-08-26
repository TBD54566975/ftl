import { useParams } from 'react-router-dom'

export const DeclPanel = () => {
  const { moduleName, declCase, declName } = useParams()

  return (
    <div className='flex-1 py-2 px-4'>
      <p>{declCase} declaration: {moduleName}.{declName}</p>
    </div>
  )
}
