import { XMarkIcon } from '@heroicons/react/24/outline'

interface Props {
  onClick: () => void
}

export const CloseButton = ({ onClick }: Props) => {
  return (
    <button type='button' onClick={onClick} className='rounded-sm hover:bg-gray-200 dark:hover:bg-slate-500'>
      <XMarkIcon className={'h-5 w-5'} />
    </button>
  )
}
