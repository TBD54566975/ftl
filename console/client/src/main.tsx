import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App.tsx'
import './index.css'
import ModulesProvider from './providers/modules-provider.tsx'
import SchemaProvider from './providers/schema-provider.tsx'

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <SchemaProvider>
        <ModulesProvider>
          <App />
        </ModulesProvider>
      </SchemaProvider>
    </BrowserRouter>
  </React.StrictMode>
)
