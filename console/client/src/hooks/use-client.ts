import { ServiceType } from '@bufbuild/protobuf'
import { PromiseClient, createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { useMemo } from 'react'

const transport = createConnectTransport({
  baseUrl: window.location.origin,
})

export const createClient = <T extends ServiceType>(service: T): PromiseClient<T> => {
  return createPromiseClient(service, transport)
}

export const useClient = <T extends ServiceType>(service: T): PromiseClient<T> => {
  return useMemo(() => createClient(service), [service])
}
