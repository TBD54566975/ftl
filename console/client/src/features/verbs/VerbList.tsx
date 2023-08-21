import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbCard } from './VerbCard'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import {  getVerbCode } from './verb.utils'

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
          <VerbCard verb={verb} />
          <div className='text-sm pt-4'>
            <SyntaxHighlighter language={module?.language || 'go'}
              style={atomDark}
            >
              {getVerbCode(verb?.verb)}
            </SyntaxHighlighter>
          </div>
        </div>
      ))}
    </>
  )
}
