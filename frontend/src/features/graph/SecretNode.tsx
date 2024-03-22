import { NodeProps } from 'reactflow'

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const SecretNode = ({ data }: Props) => {
  return (
    <>
      <div className={`grid h-full w-full bg-green-600 rounded-md ${data.selected ? 'bg-pink-600' : ''}`}>
        <div className='place-self-center text-xs text-gray-100 truncate max-w-[90%]'>{data.title}</div>
      </div>
    </>
  )
}
