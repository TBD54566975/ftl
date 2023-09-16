import { GetModulesResponse } from '../../protos/xyz/block/ftl/v1/console/console_pb'

interface Call {
  module: string
  name: string
}

interface VerbItem {
  name?: string
  'data-id': string
  calls: Call[]
}

interface Item {
  'data-id': string
  name: string
  style: { marginLeft: number }
  verbs: VerbItem[]
}

export const createLayoutDataStructure = (data: GetModulesResponse): Item[] => {
  const graph: { [key: string]: Set<string> } = {}
  const items: Item[] = []

  // Initialize graph with all module names
  data.modules.forEach((module) => {
    graph[module.name] = new Set()
  })

  // Populate graph with relationships based on verbs' metadata
  data.modules.forEach((module) => {
    module.verbs.forEach((verbEntry) => {
      const verb = verbEntry.verb
      verb?.metadata.forEach((metadataEntry) => {
        if (metadataEntry.value.case === 'calls') {
          metadataEntry.value.value.calls.forEach((call) => {
            if (call.module) {
              graph[module.name].add(call.module)
            }
          })
        }
      })
    })
  })

  // Helper function to determine depth of a node in the graph
  const determineDepth = (
    node: string,
    visited: Set<string> = new Set(),
    ancestors: Set<string> = new Set(),
  ): number => {
    if (ancestors.has(node)) {
      // Cycle detected
      return 0
    }

    let depth = 0
    ancestors.add(node)
    graph[node].forEach((neighbor) => {
      if (!visited.has(neighbor)) {
        visited.add(neighbor)
        depth = Math.max(depth, 1 + determineDepth(neighbor, visited, ancestors))
      }
    })
    ancestors.delete(node)

    return depth
  }

  // Create the new structure
  Object.keys(graph)
    .sort()
    .forEach((moduleName) => {
      const moduleData = data.modules.find((mod) => mod.name === moduleName)
      if (!moduleData) return

      const depth = determineDepth(moduleName)
      const item: Item = {
        'data-id': moduleName,
        name: moduleName,
        style: { marginLeft: 20 * (depth + 1) },
        verbs: [],
      }

      moduleData.verbs.forEach((verbEntry) => {
        const verb = verbEntry.verb
        const verbItem: VerbItem = {
          name: verb?.name,
          'data-id': `${moduleName}.${verb?.name}`,
          calls: [],
        }
        verb?.metadata.forEach((metadataEntry) => {
          if (metadataEntry.value.case === 'calls') {
            metadataEntry.value.value.calls.forEach((call) => {
              verbItem.calls.push({
                module: call.module,
                name: call.name,
              })
            })
          }
        })
        item.verbs.push(verbItem)
      })

      items.push(item)
    })

  return items
}
