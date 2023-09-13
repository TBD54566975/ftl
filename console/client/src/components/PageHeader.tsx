import { panelColor } from '../utils'

interface Props {
  icon?: React.ReactNode
  title?: string
  children?: React.ReactNode
}

export const PageHeader = ({ icon, title, children }: Props) => {
  return (
    <div
      className={`sticky top-0 z-10 ${panelColor} shadow dark:shadow-md flex justify-between items-center py-2 px-4 text-gray-70`}
    >
      <div className='flex items-center'>
        <span className='-mt-0.5 text-indigo-500 pr-1'>{icon}</span>
        <span className='text-lg'>{title}</span>
      </div>
      {children}
    </div>
  )
}
