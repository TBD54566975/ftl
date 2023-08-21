import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App.tsx'
import './index.css'
import ModulesProvider from './providers/modules-provider.tsx'
import SchemaProvider from './providers/schema-provider.tsx'
import { SelectedModuleProvider } from './providers/selected-module-provider.tsx'
import { SelectedTimelineEntryProvider } from './providers/selected-timeline-entry-provider.tsx'
import { TabsProvider } from './providers/tabs-provider.tsx'

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <SchemaProvider>
        <ModulesProvider>
          <SelectedModuleProvider>
            <SelectedTimelineEntryProvider>
              <TabsProvider>
                <App />
              </TabsProvider>
            </SelectedTimelineEntryProvider>
          </SelectedModuleProvider>
        </ModulesProvider>
      </SchemaProvider>
    </BrowserRouter>
  </React.StrictMode>
)
