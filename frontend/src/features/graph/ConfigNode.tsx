import type { NodeProps } from 'reactflow'

export const configHeight = 24

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const ConfigNode = ({ data }: Props) => {
  return (
    <>
      <div className={`grid h-full w-full  rounded-md ${data.selected ? 'bg-pink-600' : 'bg-slate-500'}`}>
        <div className='place-self-center text-xs text-gray-100 truncate max-w-[90%]'>{data.title}</div>
      </div>
    </>
  )
}
