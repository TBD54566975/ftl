interface Props {
  children: React.ReactNode
  onClick?: () => void
  className?: string
}

export const ButtonExtraSmall = ({ children, onClick }: Props) => {
  return (
    <button
      type='button'
      onClick={onClick}
      className='rounded bg-indigo-600 px-1.5 py-0.5 text-xs font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
    >
      {children}
    </button>
  )
}
