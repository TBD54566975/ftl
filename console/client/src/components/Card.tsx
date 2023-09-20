interface Props {
  children: React.ReactNode
}
export const Card = ({ children }: Props) => {
  return <div className='p-2 rounded-md border border-gray-500'>{children}</div>
}
