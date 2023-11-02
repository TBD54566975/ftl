interface Props {
  children: React.ReactNode
  onClick?: () => void
  className?: string
}

export const ButtonSmall = ({ children, onClick }: Props) => {
  return (
    <button
      type='button'
      onClick={onClick}
      className='rounded bg-indigo-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
    >
      {children}
    </button>
  )
}
