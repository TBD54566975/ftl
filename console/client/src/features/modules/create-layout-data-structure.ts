import { GetModulesResponse } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { Node, Position } from 'reactflow'
interface Call {
  module: string
  name: string
}

type VerbID = `${string}.${string}`
export interface VerbItem {
  name?: string
  id: VerbID
  calls: Call[]
}

export interface Item {
  name: string
  verbs: VerbItem[]
  depth: number
}

type ModuleMap = Map<number, Set<Item>>
interface Graph {
  [key: string]: Set<string>
}

type SourceId = `${string}.${string}-source`
type TargetId = `${string}.${string}-source`

type ModuleNode = Node<Item>

interface Edge {
  id: string
  source: string
  target: string
  sourceHandle: SourceId
  targetHandle: TargetId
  animated: true
  type: 'smoothstep'
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

const nodePositionYDefault = 150
const nodePositionXDefault = 250

export const createLayoutDataStructure = (data: GetModulesResponse): [ModuleNode[], Edge[]] => {
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
  const depthMap: Map<number, Set<Item>> = new Map()

  sortedKeys.forEach((moduleName) => {
    const moduleData = data.modules.find((mod) => mod.name === moduleName)
    if (!moduleData) return

    const depth = determineDepth(moduleName)

    const item: Item = {
      name: moduleName,
      verbs: [],
      depth,
    }

    moduleData.verbs.forEach((verbEntry) => {
      const verb = verbEntry.verb
      const verbItem: VerbItem = {
        name: verb?.name,
        id: `${moduleName}.${verb?.name}`,
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
    depthMap.has(depth) ? depthMap.get(depth)?.add(item) : depthMap.set(depth, new Set([item]))
  })
  // Sorted Modules
  const sortedModules = flattenMap(depthMap, graph)

  const nodes: Node[] = []
  const edges: Edge[] = []
  let y = 0
  for (const module of sortedModules) {
    nodes.push({
      type: 'moduleNode',
      id: module.name,
      sourcePosition: Position.Right,
      targetPosition: Position.Left,
      position: { x: nodePositionXDefault * module.depth, y },
      data: module,
    })
    y = nodePositionYDefault + y
    module.verbs.forEach(({ id, calls }) => {
      if (calls.length) {
        edges.push(
          ...calls.map<Edge>((call): Edge => {
            const targetId = `${call.module}.${call.name}`
            return {
              id: `${id}-${targetId}`,
              source: module.name,
              target: call.module,
              sourceHandle: `${id}-source` as SourceId,
              targetHandle: `${targetId}-target` as TargetId,
              animated: true,
              type: 'smoothstep',
            }
          }),
        )
      }
    })
  }

  return [nodes, edges]
}
