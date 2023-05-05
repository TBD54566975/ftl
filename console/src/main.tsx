import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App.tsx'
import './index.css'
import SchemaProvider from './providers/schema-provider.tsx'

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <SchemaProvider>
        <App />
      </SchemaProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
