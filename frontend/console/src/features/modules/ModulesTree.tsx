import { Disclosure, DisclosureButton, DisclosurePanel } from '@headlessui/react'
import {
  AnonymousIcon,
  ArrowRight01Icon,
  BubbleChatIcon,
  CircleArrowRight02Icon,
  CodeIcon,
  DatabaseIcon,
  FileExportIcon,
  FlowIcon,
  FunctionIcon,
  type HugeiconsProps,
  LeftToRightListNumberIcon,
  MessageIncoming02Icon,
  PackageIcon,
  Settings02Icon,
  SquareLock02Icon,
} from 'hugeicons-react'
import { useEffect, useMemo, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { Decl } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { classNames } from '../../utils'
import type { ModuleTreeItem } from './module.utils'
import { addModuleToLocalStorageIfMissing, listExpandedModulesFromLocalStorage, toggleModuleExpansionInLocalStorage } from './module.utils'

const ExportedIcon = () => <FileExportIcon className='size-4 text-indigo-500 -ml-1' />

type IconMap = Record<string, React.FC<Omit<HugeiconsProps, 'ref'> & React.RefAttributes<SVGSVGElement>>>
const icons: IconMap = {
  config: Settings02Icon,
  data: CodeIcon,
  database: DatabaseIcon,
  enum: LeftToRightListNumberIcon,
  fsm: FlowIcon,
  topic: BubbleChatIcon,
  typeAlias: AnonymousIcon,
  secret: SquareLock02Icon,
  subscription: MessageIncoming02Icon,
  verb: FunctionIcon,
}

type WithExport = { export?: boolean }

const DeclNode = ({ decl, href, isSelected }: { decl: Decl; href: string; isSelected: boolean }) => {
  if (!decl.value || !decl.value.case || !decl.value.value) {
    return []
  }
  const navigate = useNavigate()
  const Icon = useMemo(() => icons[decl.value.case || ''] || CodeIcon, [decl.value.case])
  return (
    <li className='my-1'>
      <div
        id={`decl-${decl.value.value.name}`}
        className={classNames(
          isSelected ? 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 hover:dark:bg-gray-600' : 'hover:bg-gray-200 hover:dark:bg-gray-700',
          'group flex items-center gap-x-2 rounded-md pl-4 pr-2 text-sm font-light leading-6 w-full cursor-pointer',
        )}
        onClick={(e) => {
          e.preventDefault()
          navigate(href)
        }}
      >
        <Icon aria-hidden='true' className='size-4 shrink-0' />
        {decl.value.value.name}
        {(decl.value.value as WithExport).export === true ? <ExportedIcon /> : []}
      </div>
    </li>
  )
}

const ModuleSection = ({ module, isExpanded, toggleExpansion }: { module: ModuleTreeItem; isExpanded: boolean; toggleExpansion: (m: string) => void }) => {
  const { moduleName, declName } = useParams()
  const navigate = useNavigate()
  const isSelected = useMemo(() => moduleName === module.name, [moduleName, module.name])
  const selectedRef = useRef<HTMLButtonElement>(null)
  const refProp = isSelected ? { ref: selectedRef } : {}

  // Scroll to the selected module on the first page load
  useEffect(() => selectedRef.current?.scrollIntoView(), [])

  return (
    <li key={module.name} id={`module-tree-module-${module.name}`} className='my-2'>
      <Disclosure as='div' defaultOpen={isExpanded}>
        <DisclosureButton
          {...refProp}
          className={classNames(
            isSelected ? 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 hover:dark:bg-gray-600' : 'hover:bg-gray-200 hover:dark:bg-gray-700',
            'group flex w-full modules-center gap-x-2 space-y-1 rounded-md px-2 text-left text-sm font-medium leading-6',
          )}
          onClick={() => toggleExpansion(module.name)}
        >
          <PackageIcon aria-hidden='true' className='size-4 my-1 shrink-0 ' />
          {module.name}
          <CircleArrowRight02Icon
            className='size-4 shrink-0 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600'
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              navigate(`/modules/${module.name}`)
            }}
          />
          {module.decls.length === 0 || (
            <ArrowRight01Icon aria-hidden='true' className='ml-auto h-4 w-4 shrink-0 group-data-[open]:rotate-90 group-data-[open]:text-gray-500' />
          )}
        </DisclosureButton>
        <DisclosurePanel as='ul' className='px-2'>
          {module.decls.map((d, i) => (
            <DeclNode
              key={i}
              decl={d}
              href={`/modules/${module.name}/${d.value.case}/${d.value.value?.name}`}
              isSelected={isSelected && declName === d.value.value?.name}
            />
          ))}
        </DisclosurePanel>
      </Disclosure>
    </li>
  )
}

export const ModulesTree = ({ modules }: { modules: ModuleTreeItem[] }) => {
  const { moduleName } = useParams()
  useEffect(() => {
    addModuleToLocalStorageIfMissing(moduleName)
  }, [moduleName])

  modules.sort((m1, m2) => Number(m1.isBuiltin) - Number(m2.isBuiltin))

  const expandedModules = listExpandedModulesFromLocalStorage()
  return (
    <div className='flex grow flex-col h-full gap-y-5 overflow-y-auto bg-gray-100 dark:bg-gray-900 px-6'>
      <nav className='flex flex-1 flex-col'>
        <ul className='flex flex-1 flex-col gap-y-7'>
          <li>
            <ul className='-mx-2'>
              {modules.map((m) => (
                <ModuleSection key={m.name} module={m} isExpanded={expandedModules.includes(m.name)} toggleExpansion={toggleModuleExpansionInLocalStorage} />
              ))}
            </ul>
          </li>
        </ul>
      </nav>
    </div>
  )
}
