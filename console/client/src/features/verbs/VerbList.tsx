import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbCard } from './VerbCard'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import {  getVerbCode } from './verb.utils'
import { EllipsisVerticalIcon } from '@heroicons/react/24/solid'
import { Modal } from '../../components/modal'
import { VerbPage } from './VerbPage'
type Props = {
  module?: Module
}

export const VerbList: React.FC<Props> = ({ module }) => {
  if(!module) return <></>
  const verbs = module?.verbs
  return (
    <>
      {verbs?.map(verb => (
        <div key={verb.verb?.name}>
          <VerbCard 
            module={module}
            verb={verb}
          />
          <div className='text-sm pt-4'>
            <SyntaxHighlighter language={module?.language || 'go'}
              style={atomDark}
            >
              {getVerbCode(verb?.verb)}
            </SyntaxHighlighter>
            <Modal
              title={verb.verb?.name}
              key={verb.verb?.name}
              trigger={({ onClick }) =>  ( <button onClick={onClick}><EllipsisVerticalIcon className='text-white h-6 w-6' /></button>)}
            >
              <VerbPage 
                module={module}
                id={verb.verb?.name}
              />
            </Modal>
          </div>
        </div>
      ))}
    </>
  )
}
