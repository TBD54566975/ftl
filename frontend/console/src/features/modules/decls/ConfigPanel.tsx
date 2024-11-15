import { useContext, useEffect, useState } from 'react'
import { Button } from '../../../components/Button'
import { CodeEditor } from '../../../components/CodeEditor'
import { ResizablePanels } from '../../../components/ResizablePanels'
import { useClient } from '../../../hooks/use-client'
import { ConsoleService } from '../../../protos/xyz/block/ftl/v1/console/console_connect'
import type { Config } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { NotificationType, NotificationsContext } from '../../../providers/notifications-provider'
import { DeclDefaultPanels } from './DeclDefaultPanels'
import { PanelHeader } from './PanelHeader'

export const ConfigPanel = ({ value, schema, moduleName, declName }: { value: Config; schema: string; moduleName: string; declName: string }) => {
  const client = useClient(ConsoleService)
  const [configValue, setConfigValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const notification = useContext(NotificationsContext)

  useEffect(() => {
    handleGetConfig()
  }, [moduleName, declName])

  const handleGetConfig = () => {
    setIsLoading(true)
    client.getConfig({ module: moduleName, name: declName }).then((resp) => {
      setConfigValue(new TextDecoder().decode(resp.value))
      setIsLoading(false)
    })
  }

  const handleSetConfig = () => {
    setIsLoading(true)
    client
      .setConfig({
        module: moduleName,
        name: declName,
        value: new TextEncoder().encode(configValue),
      })
      .then(() => {
        setIsLoading(false)
        notification?.showNotification({
          title: 'Config updated',
          message: 'Config updated successfully',
          type: NotificationType.Success,
        })
      })
      .catch((error) => {
        setIsLoading(false)
        notification?.showNotification({
          title: 'Failed to update config',
          message: error.message,
          type: NotificationType.Error,
        })
      })
  }

  if (!value || !schema) {
    return null
  }
  const decl = value.config
  if (!decl) {
    return null
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='Config' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
              <CodeEditor value={configValue} onTextChanged={setConfigValue} />
              <div className='mt-2 space-x-2 flex flex-nowrap justify-end'>
                <Button onClick={handleSetConfig} disabled={isLoading}>
                  Save
                </Button>
                <Button onClick={handleGetConfig} disabled={isLoading}>
                  Refresh
                </Button>
              </div>
            </div>
          </div>
        }
        rightPanelHeader={undefined}
        rightPanelPanels={DeclDefaultPanels(schema, value.references)}
        storageKeyPrefix='configPanel'
      />
    </div>
  )
}
