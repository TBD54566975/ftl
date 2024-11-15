import { Badge } from '../../../components/Badge'

export const PanelHeader = ({ title, declRef, exported, comments }: { title?: string; declRef?: string; exported: boolean; comments?: string[] }) => {
  return (
    <div className='mb-4 pb-2 border-b border-gray-200 dark:border-gray-600'>
      <div className='flex items-center space-x-4 justify-between'>
        <span className='text-xl font-semibold'>{title}</span>
        {exported && (
          <div className='mb-2'>
            <Badge name='Exported' />
          </div>
        )}
      </div>
      {declRef && <span className='font-roboto-mono text-sky-500 text-md'>{declRef}</span>}
      {comments && comments.length > 0 && <p className='text-sm italic my-1'>{comments.join(' ')}</p>}
    </div>
  )
}
