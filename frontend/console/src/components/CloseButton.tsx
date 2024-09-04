import { Cancel01Icon } from 'hugeicons-react'

interface Props {
  onClick: () => void
}

export const CloseButton = ({ onClick }: Props) => {
  return (
    <button type='button' onClick={onClick} className='rounded-sm hover:bg-gray-200 dark:hover:bg-slate-500'>
      <Cancel01Icon className={'h-5 w-5'} />
    </button>
  )
}
