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

export default function ModulesPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const schema = React.useContext(schemaContext)
  const modules = React.useContext(modulesContext)
  const [ searchParams ] = useSearchParams()
  const id = searchParams.get('module')
  const [ name, setName ] = React.useState(id || undefined)
  const module = modules.modules.find(module => module?.name === id)
  const handleChange = (value: string) =>{
    if(value === '') {
      setName(undefined) 
      searchParams.delete('module')
    } else {
      searchParams.set('module', value)
      setName(value)
    } 
    navigate({ ...location, search: searchParams.toString() })
  }
  return (
    <div className={styles.grid}>
      <div className={ styles.filter}>
        <Disclosure defaultOpen={true}>
          <RadioGroup onChange={handleChange}>
            <Disclosure.Button className='py-2'>
              <RadioGroup.Label>Modules</RadioGroup.Label>
            </Disclosure.Button>
            <Disclosure.Panel className={styles.modules}>
              <RadioGroup.Option value=''
                defaultChecked
              >
                {({ checked }) => (
                  <div
                    className={styles.radio}
                    key={name}
                  >
                    <div>
                      All
                    </div>
                  </div>
                )}
              </RadioGroup.Option>
              {schema.map(module => {
                const name = module.schema?.name
                return (
                  <RadioGroup.Option value={name}
                    key={name}
                  >
                    {({ checked }) => (
                      <div
                        className={styles.radio}
                        key={name}
                        onClick={() => {
                    
                        }}
                      >
                        <div>
                          {module.deploymentName}
                        </div>
                        {(module.schema?.comments.length ?? 0) > 0 && (
                          <div className={styles.radioComment}>{module.schema?.comments}</div>
                        )}
                      </div>
                    )}
                  </RadioGroup.Option>
                
                )})}
            </Disclosure.Panel>
          </RadioGroup>  
        </Disclosure>
        {module && <Disclosure>
          <Disclosure.Button className='py-2'>
            Verbs: {module.name}            
          </Disclosure.Button>
          <Disclosure.Panel className={styles.modules}>
            <div className={styles.misc}>
              <VerbList module={module} />
            </div>
          </Disclosure.Panel>
        </Disclosure>}
      </div>
      <Timeline module={module} />
      <RequestModal />
      <VerbModal />
    </div>
  )
}
