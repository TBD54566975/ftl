import React from 'react'

export const Sidebar:React.FC<{ className: string}> = ({ className }) => {
   const {modules} = React.useContext(modulesContext)
  const [ selected, setSelected ] = React.useState()
  return <div className={className}>sidebar</div>
}