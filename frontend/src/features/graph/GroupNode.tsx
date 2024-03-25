import { NodeProps } from 'reactflow'

export const groupPadding = 40

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const GroupNode = ({ data }: Props) => {
  return (
    <>
      <div
        className={`h-full rounded-md ${data.selected ? 'bg-opacity-80 bg-pink-600' : 'bg-indigo-900 bg-opacity-30'}`}
      >
        <div className='flex justify-center text-xs text-gray-100 pt-3 pl-5 truncate max-w-[90%]'>{data.title}</div>
      </div>
    </>
  )
}
