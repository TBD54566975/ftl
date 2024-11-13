import { describe, it } from '@jest/globals'
import { act, render } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { App } from './App'
import { ListResponse } from './api/productcatalog'

describe('App', () => {
  it('renders', async () => {
    const mockResponse: ListResponse = {
      products: [],
    }
    jest.spyOn(global, 'fetch').mockResolvedValueOnce({
      json: async () => mockResponse,
      ok: true,
    } as Response) // fixed the type error

    await act(async () => {
      render(
        <BrowserRouter>
          <App />
        </BrowserRouter>,
      )
    })
  })
})
