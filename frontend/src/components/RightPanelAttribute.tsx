export const RightPanelAttribute = ({ name, value }: { name?: string, value?: string }) => {
  return (
    <div className='flex justify-between space-x-2 items-center text-sm'>
      <span className='text-gray-500 dark:text-gray-400'>{name}</span>
      <span className='flex-1 min-w-0 text-right' title={value}>
        <pre className='whitespace-pre-wrap overflow-x-scroll'>{value}</pre>
      </span>
    </div>
  )
}
