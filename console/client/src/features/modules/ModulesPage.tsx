import React from 'react'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'
import { Disclosure } from '@headlessui/react'
import { createLayoutDataStructure } from './create-layout-data-structure'
import { useP5Lines } from './use-p5-lines.ts'

export const ModulesPage = () => {
  const modules = React.useContext(modulesContext)
  const data = createLayoutDataStructure(modules)
  const [container, setContainer] = React.useState<HTMLDivElement>()
  const ref = React.useRef(null)
  useP5Lines({ data, container })
  React.useEffect(() => {
    const current = ref.current
    current && setContainer(current)
  }, [])
  return (
    <div className='h-full w-full flex flex-col'>
      <PageHeader icon={<Square3Stack3DIcon />} title='Modules' />
      <div ref={ref} className='flex-1 relative p-8'>
        <div role='list' className='inline-flex flex-col space-y-3 p-2'>
          {data?.map(({ name, style, verbs, 'data-id': dataId }) => (
            <Disclosure
              as='div'
              key={name}
              style={{ ...style }}
              className='min-w-fit w-44 border border-gray-100 dark:border-slate-700 rounded overflow-hidden inline-block'
            >
              <Disclosure.Button
                data-id={dataId}
                as='button'
                className='text-gray-600 dark:text-gray-300 p-1 w-full text-left flex justify-between items-center'
              >
                {name}
              </Disclosure.Button>
              <Disclosure.Panel as='ul' className='text-gray-400 dark:text-gray-400 text-xs p-1 space-y-1 list-inside'>
                {verbs.map(({ name, 'data-id': dataId }) => (
                  <li key={name} className='flex items-center text-gray-900 dark:text-gray-400' data-id={dataId}>
                    <svg
                      className='w-3.5 h-3.5 mr-2 text-gray-500 dark:text-gray-400 flex-shrink-0'
                      aria-hidden='true'
                      xmlns='http://www.w3.org/2000/svg'
                      fill='currentColor'
                      viewBox='0 0 20 20'
                    >
                      <circle cx='10' cy='10' r='4.5' />
                    </svg>
                    {name}
                  </li>
                ))}
              </Disclosure.Panel>
            </Disclosure>
          ))}
        </div>
      </div>
    </div>
  )
}
