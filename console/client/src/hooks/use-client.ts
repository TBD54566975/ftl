import { PromiseClient, createPromiseClient } from '@bufbuild/connect'
import { createConnectTransport } from '@bufbuild/connect-web'
import { ServiceType } from '@bufbuild/protobuf'
import { useMemo } from 'react'

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8892',
})

export const createClient = <T extends ServiceType>(service: T): PromiseClient<T> => {
  return createPromiseClient(service, transport)
}

export const useClient = <T extends ServiceType>(service: T): PromiseClient<T> => {
  return useMemo(() => createClient(service), [service])
}
