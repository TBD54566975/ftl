import { NodeProps } from 'reactflow'

interface Props extends NodeProps {
  data: {
    title: string
  }
}

export const GroupNode = ({ data }: Props) => {
  return (
    <>
      <div className='h-full bg-indigo-800 rounded-md'>
        <div className='flex justify-center text-xs text-gray-100 pt-2'>{data.title}</div>
      </div>
    </>
  )
}
