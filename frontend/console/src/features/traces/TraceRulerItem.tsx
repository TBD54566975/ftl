import { TraceGraphRuler } from './TraceGraphRuler'

export const TraceRulerItem = ({ duration }: { duration: number }) => {
  return (
    <li key='trace-ruler-item' className='flex items-center justify-between px-2'>
      <span className='flex items-center w-1/2 text-sm gap-x-2 font-medium' />
      <div className='relative w-2/3 h-full flex-grow'>
        <TraceGraphRuler duration={duration} />
      </div>
      <span className='text-xs font-medium ml-4 w-20 text-right' />
    </li>
  )
}
