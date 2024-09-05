import type { ReactNode } from 'react'
import { Badge } from '../../../components/Badge'

export const PanelHeader = ({ children, exported, comments }: { children?: ReactNode; exported: boolean; comments?: string[] }) => {
  return (
    <div className='flex-1'>
      {exported && (
        <div className='mb-2'>
          <Badge name='Exported' />
        </div>
      )}
      {children}
      {comments && comments.length > 0 && <p className='text-xs my-1'>{comments}</p>}
    </div>
  )
}
