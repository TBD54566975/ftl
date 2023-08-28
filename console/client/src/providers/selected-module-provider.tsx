import {PropsWithChildren, createContext, useState} from 'react'
import {Module} from '../protos/xyz/block/ftl/v1/console/console_pb'

type SelectedModuleContextType = {
  selectedModule: Module | null
  setSelectedModule: React.Dispatch<React.SetStateAction<Module | null>>
}

export const SelectedModuleContext = createContext<SelectedModuleContextType>({
  selectedModule: null,
  setSelectedModule: () => {},
})

export const SelectedModuleProvider = (props: PropsWithChildren) => {
  const [selectedModule, setSelectedModule] = useState<Module | null>(null)

  return (
    <SelectedModuleContext.Provider value={{selectedModule, setSelectedModule}}>
      {props.children}
    </SelectedModuleContext.Provider>
  )
}
