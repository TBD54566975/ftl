import { CubeTransparentIcon } from '@heroicons/react/24/outline'
import { Page } from '../../layout'
import { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { GraphPane } from './GraphPane'

export type FTLNode = Module | Verb | Secret | Config

export const GraphPage = () => {
  return (
    <Page>
      <Page.Header icon={<CubeTransparentIcon />} title='Graph' />
      <Page.Body className='flex h-full bg-slate-800'>
        <GraphPane />
      </Page.Body>
    </Page>
  )
}
