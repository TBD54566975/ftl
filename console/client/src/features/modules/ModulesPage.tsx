import React from 'react'
import { schemaContext } from '../../providers/schema-provider.tsx'
import { useSearchParams, useNavigate , useLocation  } from 'react-router-dom'
import { modulesContext } from '../../providers/modules-provider.tsx'
import * as styles from './Module.module.css'
import { Timeline } from '../timeline/Timeline.tsx'
import { VerbList } from '../verbs/VerbList.tsx'
import { Disclosure, RadioGroup } from '@headlessui/react'
import { RequestModal } from '../requests/RequestsModal.tsx'
import { VerbModal } from '../verbs/VerbModal.tsx'
import { ChevronUpIcon, CheckIcon } from '@heroicons/react/20/solid'

export default function ModulesPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const schema = React.useContext(schemaContext)
  const modules = React.useContext(modulesContext)
  const [ searchParams ] = useSearchParams()
  const id = searchParams.get('module')
  const module = modules.modules.find(module => module?.name === id)
  const handleChange = (value: string) =>{
    if(value === '') {
      searchParams.delete('module')
    } else {
      searchParams.set('module', value)
    } 
    navigate({ ...location, search: searchParams.toString() })
  }
  if(schema.length === 0) return <></>
  return (
    <div className={styles.grid}>
      <div className={`
        mx-auto
        w-full
        max-w-md
        rounded-2xl
        bg-indigo-50
        dark:bg-slate-700
        p-2
        flex
        flex-col
        gap-3
        `}
      >
        <Disclosure defaultOpen={true}>
          {({ open }) => (  <RadioGroup onChange={handleChange}>
            <Disclosure.Button className={`flex w-full justify-between rounded-lg bg-indigo-600 px-4 py-2 text-left text-sm font-medium text-white hover:bg-indigo-400 focus:outline-none focus-visible:ring focus-visible:ring-purple-500 focus-visible:ring-opacity-75`}>
              Modules
              <ChevronUpIcon
                className={`${
                    open ? 'rotate-180 transform' : ''
                } h-5 w-5 text-white`}
              />
            </Disclosure.Button>
            <Disclosure.Panel className={`px-4 pt-4 pb-2 text-sm text-gray-500`}>
              <div className='space-y-2'>
                <RadioGroup.Option
                  value={''}
                  className={({ active, checked }) =>
                    `${
                    active
                      ? 'ring-2 ring-white ring-opacity-60 ring-offset-2 ring-offset-sky-300'
                      : ''
                    }
                  ${
                    checked ? 'bg-sky-900 bg-opacity-75 text-white' : 'bg-white'
            }
                    relative
                    flex
                    cursor-pointer
                    rounded-lg
                    px-3
                    py-2
                    shadow-md
                    focus:outline-none
                    bg-white
                    dark:bg-slate-800
                    `
                  }
                >
                  {({  checked }) => (
                    <>
                      <div className='flex w-full items-center justify-between'>
                        <div className='flex items-center'>
                          <div className='text-sm'>
                            <RadioGroup.Label
                              as='p'
                              className={`font-medium  ${
                              checked ? 'text-white' : 'text-gray-900 dark:text-white'
                              }`}
                            >
                            all
                            </RadioGroup.Label>
                          </div>
                        </div>
                        {checked && (
                          <div className='shrink-0 text-white'>
                            <CheckIcon className='h-6 w-6' />
                          </div>
                        )}
                      </div>
                    </>
                  )}
                </RadioGroup.Option>
                {schema.map(module => {
                  const name = module.schema?.name
                  return (
                    <RadioGroup.Option
                      key={name}
                      value={name}
                      className={({ active, checked }) =>
                        `${
                      active
                        ? 'ring-2 ring-white ring-opacity-60 ring-offset-2 ring-offset-sky-300'
                        : ''
                        }
                    ${
                      checked ? 'bg-sky-900 bg-opacity-75 text-white' : 'bg-white'
                    }
                      relative
                      flex
                      cursor-pointer
                      rounded-lg
                      px-3
                      py-2
                      shadow-md
                      focus:outline-none
                      dark:bg-slate-800
                      `
                      }
                    >
                      {({  checked }) => (
                        <>
                          <div className='flex w-full items-center justify-between'>
                            <div className='flex items-center'>
                              <div className='text-sm'>
                                <RadioGroup.Label
                                  as='p'
                                  className={`font-medium  ${
                                checked ? 'text-white' : 'text-gray-900 dark:text-white'
                                  }`}
                                >
                                  {module.deploymentName}
                                </RadioGroup.Label>
                                {(module.schema?.comments.length ?? 0) > 0 && (<RadioGroup.Description
                                  as='span'
                                  className={`inline ${
                                checked ? 'text-sky-100' : 'text-gray-500 dark:text-white'
                                  }`}
                                >
                                  <span>{module.schema?.comments}</span>
                                </RadioGroup.Description>)}
                              </div>
                            </div>
                            {checked && (
                              <div className='shrink-0 text-white'>
                                <CheckIcon className='h-6 w-6' />
                              </div>
                            )}
                          </div>
                        </>
                      )}
                    </RadioGroup.Option>
                  )})}
              </div>
            </Disclosure.Panel>
          </RadioGroup>  
          )}
        
        </Disclosure>
        {module &&  <Disclosure defaultOpen={true}>
          {({ open }) => (  <RadioGroup onChange={handleChange}>
            <Disclosure.Button className={`flex w-full justify-between rounded-lg bg-indigo-600 px-4 py-2 text-left text-sm font-medium text-white hover:bg-indigo-400 focus:outline-none focus-visible:ring focus-visible:ring-purple-500 focus-visible:ring-opacity-75`}>
              {module.name}: verbs  
              <ChevronUpIcon
                className={`${
                    open ? 'rotate-180 transform' : ''
                } h-5 w-5 text-white`}
              />
            </Disclosure.Button>
            <Disclosure.Panel className={`px-4 pt-4 pb-2 text-sm text-gray-500`}>
              <VerbList module={module} />
            </Disclosure.Panel>
          </RadioGroup>  
          )}
        
        </Disclosure>}
      </div>
      <Timeline module={module} />
      <RequestModal />
      <VerbModal />
    </div>
  )
}
