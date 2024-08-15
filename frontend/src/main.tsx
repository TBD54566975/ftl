import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import { Buffer } from 'buffer'
import { AppProviders } from './providers/AppProviders.tsx'

globalThis.Buffer = Buffer

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <AppProviders />
    </BrowserRouter>
  </React.StrictMode>,
)
