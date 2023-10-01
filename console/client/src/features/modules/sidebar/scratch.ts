const [filteredVerbs, setFilteredVerbs] = React.useState<string[]>([])
const [filteredModules, setFilteredModules] = React.useState<[
  string, 
  {deploymentName: string, verbs: string[]}
][]>([...modulesMap])
const [selectedModule, setSelectedModule] = React.useState<string>()
const [showModules, setShowModules] = React.useState<boolean>(true)

const handleModuleClick: React.MouseEventHandler<HTMLButtonElement> = (evt) => {
  const name = evt.currentTarget.value
  setSelectedModule(name)
  const verbs = modulesMap.get(name)?.verbs
  verbs && setFilteredVerbs(verbs)
  setShowModules(false)
}

const handleVerbClick: React.MouseEventHandler<HTMLButtonElement> = (evt) => {
  const id = evt.currentTarget.value

}

const handleZoomTo: React.MouseEventHandler<HTMLButtonElement> = (evt) => {
  const name = evt.currentTarget.value
}

const handleFilter: React.ChangeEventHandler<HTMLInputElement> = evt => {

}

const handleShowModule: React.MouseEventHandler<HTMLButtonElement> = _ => {
  setSelectedModule(undefined)
  setShowModules(true)
}