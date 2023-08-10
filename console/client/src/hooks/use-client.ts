import { PromiseClient, createPromiseClient } from '@bufbuild/connect'
import { createConnectTransport } from '@bufbuild/connect-web'
import { ServiceType } from '@bufbuild/protobuf'
import { useMemo } from 'react'

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8892',
})

export function useClient<T extends ServiceType>(service: T): PromiseClient<T> {
  return useMemo(() => createPromiseClient(service, transport), [ service ])
}
