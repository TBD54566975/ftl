export const RightPanelHeader = ({ Icon, title }: { Icon?: React.ElementType; title?: string }) => {
  return (
    <div className='flex items-center gap-2 px-2 py-2'>
      {Icon && <Icon className='h-5 w-5 text-indigo-500' />}
      {title && <div className='flex flex-col min-w-0'>{title}</div>}
    </div>
  )
}
