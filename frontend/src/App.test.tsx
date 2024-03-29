import { describe, it } from '@jest/globals'
import { render } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { App } from './App'

describe('App', () => {
  it('renders the app', () => {
    window.history.pushState({}, 'Modules', '/modules')

    render(
      <BrowserRouter>
        <App />
      </BrowserRouter>,
    )
  })
})
