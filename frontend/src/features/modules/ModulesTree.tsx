import type { ForwardRefExoticComponent, SVGProps } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Disclosure, DisclosureButton, DisclosurePanel } from '@headlessui/react'
import { ArrowRightCircleIcon, BellIcon, BellAlertIcon, BoltIcon, BookOpenIcon, CircleStackIcon, ChevronRightIcon, Cog6ToothIcon, DocumentDuplicateIcon, LockClosedIcon, NumberedListIcon, SquaresPlusIcon } from '@heroicons/react/24/outline'
import { TableCellsIcon } from '@heroicons/react/24/solid'
import { classNames } from '../../utils'
import type { ModuleTreeItem } from './module.utils'
import type { Decl } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

// This could alternatively be an icon, but we'd need to pick a good one.
const ExportBadge = () => (
  <span className='text-xs py-0.5 px-1.5 bg-gray-200 dark:bg-gray-900 dark:text-gray-300 rounded-md'>
    Exported
  </span>
)

type IconMap = Record<string, ForwardRefExoticComponent<SVGProps<SVGSVGElement> & { title?: string; titleId?: string }>>
const icons: IconMap = {
  'config': Cog6ToothIcon,
  'data': TableCellsIcon,
  'database': CircleStackIcon,
  'enum': NumberedListIcon,
  'fsm': SquaresPlusIcon,
  'topic': BellIcon,
  'typeAlias': DocumentDuplicateIcon,
  'secret': LockClosedIcon,
  'subscription': BellAlertIcon,
  'verb': BoltIcon,
}

const DeclNode = ({ decl, href }: { decl: Decl, href: string }) => {
  if (!decl.value || !decl.value.case || !decl.value.value) {
    return []
  }
  const Icon = icons[decl.value.case] || BookOpenIcon
  return (
    <li className='my-1'>
      <DisclosureButton
        as='a'
        href={href}
        className={classNames(
          'hover:bg-gray-100 hover:dark:bg-gray-700',
          'group flex items-center gap-x-2 rounded-md pl-4 pr-2 text-sm font-light leading-6',
        )}
      >
        <Icon aria-hidden='true' className='size-4 shrink-0' />
        {decl.value.value.name}
        {(decl.value.value as any).export == true ? <ExportBadge /> : []}
      </DisclosureButton>
    </li>
  )
}

const ModuleSection = ({ module, isExpanded, path, params, toggleExpansion }: { module: ModuleTreeItem, isExpanded: boolean, path: string, params: string, toggleExpansion: (m: string) => void }) => {
  return (
    <li key={module.name} id={`module-tree-module-${module.name}`} className='my-2'>
      {module.decls.length === 0 ? (
        <a
          href={`${path}?${params}`}
          className={classNames(
            'hover:bg-gray-100 hover:dark:bg-gray-700',
            'group flex gap-x-3 rounded-md px-2 text-sm font-medium leading-6',
          )}
        >
          <BookOpenIcon aria-hidden='true' className='size-3 shrink-0' />
          {module.name}
        </a>
      ) : (
        <Disclosure as='div' defaultOpen={isExpanded}>
          <DisclosureButton
            className={classNames(
              'hover:bg-gray-100 hover:dark:bg-gray-700',
              'group flex w-full modules-center gap-x-2 space-y-1 rounded-md px-2 text-left text-sm font-medium leading-6',
            )}
            onClick={() => toggleExpansion(module.name)}
          >
            <BookOpenIcon aria-hidden='true' className='size-4 my-1 shrink-0 ' />
            {module.name}
            <a href={`${path}?${params}`}>
              <ArrowRightCircleIcon
                aria-hidden='true'
                className='size-4 shrink-0 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600'
                onClick={(e) => {e.stopPropagation()}}
              />
            </a>
            <ChevronRightIcon aria-hidden='true' className='ml-auto h-4 w-4 shrink-0 group-data-[open]:rotate-90 group-data-[open]:text-gray-500' />
          </DisclosureButton>
          <DisclosurePanel as='ul' className='px-2'>
            {module.decls.map((d, i) => (
              <DeclNode key={i} decl={d} href={`${path}/${d.value.case}/${d.value.value?.name}?${params}`} />
            ))}
          </DisclosurePanel>
        </Disclosure>
      )}
    </li>
  )
}

export const ModulesTree = ({ modules }: { modules: ModuleTreeItem[] }) => {
  const [searchParams, setSearchParams] = useSearchParams()
  modules.sort((m1, m2) => Number(m1.isBuiltin) - Number(m2.isBuiltin))

  const expandedModules = (searchParams.get('tree_m') || '').split(',')

  function toggleModuleExpansion(moduleName: string) {
    const expanded = (searchParams.get('tree_m') || '').split(',')
    const i = expanded.indexOf(moduleName)
    if (i === -1) {
      searchParams.set('tree_m', [...expanded, moduleName].join(','))
    } else {
      expanded.splice(i, 1)
      if (expanded.length === 1) {
        searchParams.delete('tree_m')
      } else {
        searchParams.set('tree_m', expanded.join(','))
      }
    }
    setSearchParams(searchParams)
  }

  return (
    <div className='flex grow flex-col h-full gap-y-5 overflow-y-auto border-r border-gray-200 dark:border-gray-600 py-2 px-6'>
      <nav className='flex flex-1 flex-col'>
        <ul className='flex flex-1 flex-col gap-y-7'>
          <li>
            <ul className='-mx-2'>
              {modules.map((m) => (
                <ModuleSection
                  key={m.name}
                  module={m}
                  isExpanded={expandedModules.includes(m.name)}
                  path={`/modules/${m.name}`}
                  params={searchParams.toString()}
                  toggleExpansion={toggleModuleExpansion}
                />))}
            </ul>
          </li>
        </ul>
      </nav>
    </div>
  )
}
