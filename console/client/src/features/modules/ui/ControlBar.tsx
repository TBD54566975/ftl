import React from 'react'

export const ControlBar: React.FC<{ className: string }> = ({className}) => {
  return (<div className={className}>{className} control bar</div>)
} 