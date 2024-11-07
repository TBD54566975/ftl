export const TraceGraphRuler = ({ duration }: { duration: number }) => {
  const tickInterval = duration / 4
  const ticks = Array.from({ length: 5 }, (_, i) => ({
    value: Math.round(i * tickInterval),
    position: `${(i * 100) / 4}%`,
  }))

  return (
    <div className='relative border-b border-gray-200 dark:border-gray-600 w-full h-6'>
      {ticks.map((tick, index) => (
        <div key={index} className='absolute bottom-0 transform -translate-x-1/2' style={{ left: tick.position }}>
          <span className='absolute bottom-2 text-xs font-roboto-mono text-gray-500 dark:text-gray-400 -translate-x-1/2 whitespace-nowrap'>{tick.value}ms</span>
          <span className='block h-2 w-[1px] bg-gray-200 dark:bg-gray-600' />
        </div>
      ))}
    </div>
  )
}
