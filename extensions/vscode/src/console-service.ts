import { createConnectTransport } from '@connectrpc/connect-web'
import { ServiceType } from '@bufbuild/protobuf'
import { PromiseClient, createPromiseClient } from '@connectrpc/connect'

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8892',
})

export const createClient = <T extends ServiceType>(service: T): PromiseClient<T> => {
  return createPromiseClient(service, transport)
}
