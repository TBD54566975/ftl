interface Props {
  name: string
  value: string
}

export const AttributeBadge = ({ name, value }: Props) => {
  return (
    <span className='inline-flex items-center px-2 py-0.5 space-x-1 rounded-md text-sm bg-gray-100 dark:bg-slate-700'>
      <span className='text-gray-500 dark:text-gray-400'>{name}</span>
      <span>:</span>
      <span className='text-gray-900 dark:text-gray-100 font-medium'>{value}</span>
    </span>
  )
}
