import { useContext, useEffect, useState } from 'react'
import { Button } from '../../../../components/Button'
import { CodeEditor } from '../../../../components/CodeEditor'
import { ResizablePanels } from '../../../../components/ResizablePanels'
import { useClient } from '../../../../hooks/use-client'
import { ConsoleService } from '../../../../protos/xyz/block/ftl/v1/console/console_connect'
import type { Secret } from '../../../../protos/xyz/block/ftl/v1/console/console_pb'
import { NotificationType, NotificationsContext } from '../../../../providers/notifications-provider'
import { declIcon } from '../../module.utils'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { secretPanels } from './SecretRightPanels'

export const SecretPanel = ({ value, schema, moduleName, declName }: { value: Secret; schema: string; moduleName: string; declName: string }) => {
  const client = useClient(ConsoleService)
  const [secretValue, setSecretValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const notification = useContext(NotificationsContext)

  useEffect(() => {
    handleGetSecret()
  }, [moduleName, declName])

  const handleGetSecret = () => {
    setIsLoading(true)
    client
      .getSecret({ module: moduleName, name: declName })
      .then((resp) => {
        setSecretValue(new TextDecoder().decode(resp.value))
        setIsLoading(false)
      })
      .catch((error) => {
        setIsLoading(false)
        notification?.showNotification({
          title: 'Failed to get secret',
          message: error.message,
          type: NotificationType.Error,
        })
      })
  }

  const handleSetSecret = () => {
    setIsLoading(true)
    client
      .setSecret({
        module: moduleName,
        name: declName,
        value: new TextEncoder().encode(secretValue),
      })
      .then(() => {
        setIsLoading(false)
        notification?.showNotification({
          title: 'Secret updated',
          message: 'Secret updated successfully',
          type: NotificationType.Success,
        })
      })
      .catch((error) => {
        setIsLoading(false)
        notification?.showNotification({
          title: 'Failed to update secret',
          message: error.message,
          type: NotificationType.Error,
        })
      })
  }

  if (!value || !schema) {
    return null
  }

  const decl = value.secret
  if (!decl) {
    return null
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='Secret' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
              <CodeEditor value={secretValue} onTextChanged={setSecretValue} />
              <div className='mt-2 space-x-2 flex flex-nowrap justify-end'>
                <Button onClick={handleSetSecret} disabled={isLoading}>
                  Save
                </Button>
                <Button onClick={handleGetSecret} disabled={isLoading}>
                  Refresh
                </Button>
              </div>
            </div>
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('secret', decl)} title={declName} />}
        rightPanelPanels={secretPanels(value, schema)}
        storageKeyPrefix='secretPanel'
      />
    </div>
  )
}
