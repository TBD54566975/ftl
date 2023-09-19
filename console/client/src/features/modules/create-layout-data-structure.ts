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

export interface Item {
  'data-id': string
  name: string
  style: { marginLeft: number }
  verbs: VerbItem[]
}

type ModuleMap = Map<number, Set<Item>>
interface Graph {
  [key: string]: Set<string>
}

const flattenMap = (map: ModuleMap, graph: Graph): Item[] => {
  const sortedKeys = Array.from(map.keys()).sort((a, b) => a - b)
  const flattenedList: Item[] = []

  for (const key of sortedKeys) {
    for (const item of map.get(key)!) {
      if (key === 0) {
        // Items with key 0 have no ancestors, so we just add them directly to the list
        flattenedList.push(item)
      } else if (graph[item.name]) {
        // Find the ancestor for the current item
        const ancestorName = Array.from(graph[item.name])[0]

        // Find the index of the ancestor in the flattenedList
        let insertionIndex = flattenedList.findIndex((i) => i.name === ancestorName)

        // If ancestor is found, find the position after the last dependent of the ancestor
        if (insertionIndex !== -1) {
          while (
            insertionIndex + 1 < flattenedList.length &&
            graph[flattenedList[insertionIndex + 1].name] &&
            Array.from(graph[flattenedList[insertionIndex + 1].name])[0] === ancestorName
          ) {
            insertionIndex++
          }
          flattenedList.splice(insertionIndex + 1, 0, item)
        } else {
          // If ancestor is not found, this is a fallback, though ideally this shouldn't happen
          flattenedList.push(item)
        }
      } else {
        // If no ancestor is found in the graph, simply push the item to the list
        flattenedList.push(item)
      }
    }
  }

  return flattenedList
}

export const createLayoutDataStructure = (data: GetModulesResponse): Item[] => {
  const graph: { [key: string]: Set<string> } = {}

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
              graph[call.module].add(module.name)
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

  const sortedKeys = Object.keys(graph).sort(new Intl.Collator().compare)
  const map: Map<number, Set<Item>> = new Map()

  sortedKeys.forEach((moduleName) => {
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
    map.has(depth) ? map.get(depth)?.add(item) : map.set(depth, new Set([item]))
  })

  return flattenMap(map, graph)
}
