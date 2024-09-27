import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useSchema } from '../schema/use-schema'

// biome-ignore lint/suspicious/noExplicitAny: support injection of getter for any decl type
export const useDecl = (declType: string, moduleName: string, declName: string, clientFn: (request: any, options?: any) => Promise<any>) => {
  const queryClient = useQueryClient()
  const { data: schemaData } = useSchema()

  const queryKey = [`${declType}_${moduleName}.${declName}`]

  useEffect(() => {
    if (schemaData) {
      queryClient.invalidateQueries({
        queryKey,
      })
    }
  }, [schemaData, queryClient])

  const fetch = async (signal: AbortSignal) => {
    try {
      console.debug(`fetching verb ${moduleName}.${declName} from FTL`)
      const decl = await clientFn({ moduleName, declName }, { signal })
      return decl
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('fetch - Connect error:', error)
        }
      } else {
        console.error('fetch:', error)
      }
      throw error
    }
  }

  return useQuery({
    queryKey,
    queryFn: async ({ signal }) => fetch(signal),
    enabled: !!schemaData,
  })
}
