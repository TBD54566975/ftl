import { ArrowRight01Icon, ArrowShrink02Icon, CircleArrowRight02Icon, FileExportIcon, PackageIcon } from 'hugeicons-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Multiselect, sortMultiselectOpts } from '../../components/Multiselect'
import type { MultiselectOpt } from '../../components/Multiselect'
import { classNames } from '../../utils'
import type { ModuleTreeItem } from './module.utils'
import {
  type DeclInfo,
  addModuleToLocalStorageIfMissing,
  collapseAllModulesInLocalStorage,
  declIcon,
  declSumTypeIsExported,
  declUrlFromInfo,
  listExpandedModulesFromLocalStorage,
  toggleModuleExpansionInLocalStorage,
} from './module.utils'
import { declTypeMultiselectOpts } from './schema/schema.utils'

const ExportedIcon = () => (
  <span className='w-4' title='Exported'>
    <FileExportIcon className='size-4 text-indigo-500 -ml-1' />
  </span>
)

const DeclNode = ({ decl, href, isSelected }: { decl: DeclInfo; href: string; isSelected: boolean }) => {
  const navigate = useNavigate()
  const declRef = useRef<HTMLDivElement>(null)

  // Scroll to the selected decl on page load
  useEffect(() => {
    if (isSelected && declRef.current) {
      const { top } = declRef.current.getBoundingClientRect()
      const { innerHeight } = window
      if (top < 64 || top > innerHeight) {
        declRef.current.scrollIntoView({ behavior: 'smooth' })
      }
    }
  }, [isSelected])

  const Icon = useMemo(() => declIcon(decl.declType), [decl.declType])
  return (
    <li className='my-1'>
      <Link
        ref={declRef}
        id={`decl-${decl.value.name}`}
        className={classNames(
          isSelected ? 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 hover:dark:bg-gray-600' : 'hover:bg-gray-200 hover:dark:bg-gray-700',
          'group flex items-center gap-x-2 pl-4 pr-2 text-sm font-light leading-6 w-full cursor-pointer scroll-mt-10',
        )}
        to={href}
      >
        <Icon aria-hidden='true' className='size-4 shrink-0 ml-3' />
        {decl.value.name}
        {declSumTypeIsExported(decl.value) ? <ExportedIcon /> : []}
      </Link>
    </li>
  )
}

const ModuleSection = ({
  module,
  isExpanded,
  toggleExpansion,
  selectedDeclTypes,
}: { module: ModuleTreeItem; isExpanded: boolean; toggleExpansion: (m: string) => void; selectedDeclTypes: MultiselectOpt[] }) => {
  const { moduleName, declName } = useParams()
  const navigate = useNavigate()
  const isSelected = useMemo(() => moduleName === module.name, [moduleName, module.name])
  const moduleRef = useRef<HTMLDivElement>(null)

  // Scroll to the selected module on page load
  useEffect(() => {
    if (isSelected && !declName && moduleRef.current) {
      const { top } = moduleRef.current.getBoundingClientRect()
      const { innerHeight } = window
      if (top < 64 || top > innerHeight) {
        moduleRef.current.scrollIntoView()
      }
    }
  }, [moduleName]) // moduleName is the selected module; module.name is the one being rendered

  const filteredDecls = useMemo(() => module.decls.filter((d) => !!selectedDeclTypes.find((o) => o.key === d.declType)), [module.decls, selectedDeclTypes])

  return (
    <li key={module.name} id={`module-tree-module-${module.name}`} className='mb-2'>
      <div
        ref={moduleRef}
        id={`module-${module.name}-tree-group`}
        className={classNames(
          isSelected ? 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 hover:dark:bg-gray-600' : 'hover:bg-gray-200 hover:dark:bg-gray-700',
          'group flex w-full modules-center gap-x-2 space-y-1 text-left text-sm font-medium cursor-pointer leading-6',
        )}
        onClick={() => toggleExpansion(module.name)}
      >
        <PackageIcon aria-hidden='true' className='size-4 my-1 ml-3 shrink-0' />
        {module.name}
        <CircleArrowRight02Icon
          id={`module-${module.name}-view-icon`}
          className='size-4 shrink-0 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600'
          onClick={(e) => {
            e.preventDefault()
            e.stopPropagation()
            navigate(`/modules/${module.name}`)
          }}
        />
        {filteredDecls.length === 0 || (
          <ArrowRight01Icon aria-hidden='true' className={`ml-auto mr-2 h-4 w-4 shrink-0 ${isExpanded ? 'rotate-90 text-gray-500' : ''}`} />
        )}
      </div>
      {isExpanded && (
        <ul>
          {filteredDecls.map((d, i) => (
            <DeclNode key={i} decl={d} href={declUrlFromInfo(module.name, d)} isSelected={isSelected && declName === d.value.name} />
          ))}
        </ul>
      )}
    </li>
  )
}

const declTypesSearchParamKey = 'dt'

export const ModulesTree = ({ modules }: { modules: ModuleTreeItem[] }) => {
  const { moduleName, declName } = useParams()

  const [searchParams, setSearchParams] = useSearchParams()
  const declTypeKeysFromUrl = searchParams.getAll(declTypesSearchParamKey)
  const declTypesFromUrl = declTypeMultiselectOpts.filter((o) => declTypeKeysFromUrl.includes(o.key))
  const [selectedDeclTypes, setSelectedDeclTypes] = useState(declTypesFromUrl.length === 0 ? declTypeMultiselectOpts : declTypesFromUrl)

  const initialExpanded = listExpandedModulesFromLocalStorage()
  const [expandedModules, setExpandedModules] = useState(initialExpanded)
  useEffect(() => {
    if (moduleName && declName) {
      addModuleToLocalStorageIfMissing(moduleName)
    }
    setExpandedModules(listExpandedModulesFromLocalStorage())
  }, [moduleName, declName])

  function msOnChange(opts: MultiselectOpt[]) {
    const params = new URLSearchParams()
    if (opts.length !== declTypeMultiselectOpts.length) {
      for (const o of sortMultiselectOpts(opts)) {
        params.append(declTypesSearchParamKey, o.key)
      }
    }
    setSearchParams(params)
    setSelectedDeclTypes(opts)
  }

  function toggle(toggledModule: string) {
    toggleModuleExpansionInLocalStorage(toggledModule)
    setExpandedModules(listExpandedModulesFromLocalStorage())
  }

  function collapseAll() {
    collapseAllModulesInLocalStorage()
    if (moduleName && declName) {
      addModuleToLocalStorageIfMissing(moduleName)
    }
    setExpandedModules(listExpandedModulesFromLocalStorage())
  }

  modules.sort((m1, m2) => Number(m1.isBuiltin) - Number(m2.isBuiltin))
  return (
    <div className='flex grow flex-col h-full gap-y-5 overflow-y-auto bg-gray-100 dark:bg-gray-900'>
      <nav>
        <div className='sticky top-0 border-b border-gray-300 bg-gray-100 dark:border-gray-800 dark:bg-gray-900 z-10'>
          <span className='block w-[calc(100%-32px)]'>
            <Multiselect allOpts={declTypeMultiselectOpts} selectedOpts={selectedDeclTypes} onChange={msOnChange} />
          </span>
          <span
            className='absolute inset-y-0 right-0 flex items-center px-1 mx-1 my-1.5 rounded-md cursor-pointer bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-100'
            onClick={collapseAll}
          >
            <ArrowShrink02Icon className='size-5 text-gray-500 dark:text-gray-300' />
          </span>
        </div>
        <ul>
          {modules.map((m) => (
            <ModuleSection
              key={m.name}
              module={m}
              isExpanded={expandedModules.includes(m.name)}
              toggleExpansion={toggle}
              selectedDeclTypes={selectedDeclTypes}
            />
          ))}
        </ul>
      </nav>
    </div>
  )
}
