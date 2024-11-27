import type { EdgeDefinition, ElementDefinition } from 'cytoscape'
import type { StreamModulesResult } from '../../api/modules/use-stream-modules'
import type { Config, Data, Database, Enum, Module, Secret, Topic, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { getNodeBackgroundColor } from './graph-styles'

export type FTLNode = Module | Verb | Secret | Config | Data | Database | Topic | Enum

const createParentNode = (module: Module, nodePositions: Record<string, { x: number; y: number }>) => ({
  group: 'nodes' as const,
  data: {
    id: module.name,
    label: module.name,
    type: 'groupNode',
    item: module,
  },
  ...(nodePositions[module.name] && {
    position: nodePositions[module.name],
  }),
})

const createChildNode = (
  parentName: string,
  childId: string,
  childLabel: string,
  childType: string,
  nodePositions: Record<string, { x: number; y: number }>,
  item: FTLNode,
  isDarkMode: boolean,
) => ({
  group: 'nodes' as const,
  data: {
    id: childId,
    label: childLabel,
    type: 'node',
    nodeType: childType,
    parent: parentName,
    item,
    backgroundColor: getNodeBackgroundColor(isDarkMode, childType),
  },
  ...(nodePositions[childId] && {
    position: nodePositions[childId],
  }),
})

const createModuleChildren = (module: Module, nodePositions: Record<string, { x: number; y: number }>, isDarkMode: boolean) => {
  const children = [
    // Create nodes for configs
    ...(module.configs || []).map((config: Config) =>
      createChildNode(module.name, nodeId(module.name, config.config?.name), config.config?.name || '', 'config', nodePositions, config, isDarkMode),
    ),
    // Create nodes for data
    ...(module.data || []).map((data: Data) =>
      createChildNode(module.name, nodeId(module.name, data.data?.name), data.data?.name || '', 'data', nodePositions, data, isDarkMode),
    ),
    // Create nodes for databases
    ...(module.databases || []).map((database: Database) =>
      createChildNode(
        module.name,
        nodeId(module.name, database.database?.name),
        database.database?.name || '',
        'database',
        nodePositions,
        database,
        isDarkMode,
      ),
    ),
    // Create nodes for enums
    ...(module.enums || []).map((enumDecl: Enum) =>
      createChildNode(module.name, nodeId(module.name, enumDecl.enum?.name), enumDecl.enum?.name || '', 'enum', nodePositions, enumDecl, isDarkMode),
    ),
    // Create nodes for secrets
    ...(module.secrets || []).map((secret: Secret) =>
      createChildNode(module.name, nodeId(module.name, secret.secret?.name), secret.secret?.name || '', 'secret', nodePositions, secret, isDarkMode),
    ),
    // Create nodes for topics
    ...(module.topics || []).map((topic: Topic) =>
      createChildNode(module.name, nodeId(module.name, topic.topic?.name), topic.topic?.name || '', 'topic', nodePositions, topic, isDarkMode),
    ),
    // Create nodes for verbs
    ...(module.verbs || []).map((verb: Verb) =>
      createChildNode(module.name, nodeId(module.name, verb.verb?.name), verb.verb?.name || '', 'verb', nodePositions, verb, isDarkMode),
    ),
  ]
  return children
}

const createChildEdge = (sourceModule: string, sourceVerb: string, targetModule: string, targetVerb: string) => ({
  group: 'edges' as const,
  data: {
    id: `edge-${nodeId(sourceModule, sourceVerb)}->${nodeId(targetModule, targetVerb)}`,
    source: nodeId(sourceModule, sourceVerb),
    target: nodeId(targetModule, targetVerb),
    type: 'childConnection',
  },
})

const createModuleEdge = (sourceModule: string, targetModule: string) => ({
  group: 'edges' as const,
  data: {
    id: `module-${sourceModule}->${targetModule}`,
    source: nodeId(sourceModule),
    target: nodeId(targetModule),
    type: 'moduleConnection',
  },
})

const createEdges = (modules: Module[]) => {
  const edges: EdgeDefinition[] = []
  const moduleConnections = new Set<string>() // Track unique module connections

  for (const module of modules) {
    // For each verb in the module
    for (const verb of module.verbs || []) {
      // For each reference in the verb
      for (const ref of verb.references || []) {
        // Only create verb-to-verb child edges
        edges.push(createChildEdge(ref.module, ref.name, module.name, verb.verb?.name || ''))

        // Track module-to-module connection for all reference types
        const [sourceModule, targetModule] = [module.name, ref.module].sort()
        moduleConnections.add(`${sourceModule}-${targetModule}`)
      }
    }

    for (const config of module.configs || []) {
      // For each reference in the verb
      for (const ref of config.references || []) {
        // Only create verb-to-verb child edges
        edges.push(createChildEdge(ref.module, ref.name, module.name, config.config?.name || ''))

        // Track module-to-module connection for all reference types
        const [sourceModule, targetModule] = [module.name, ref.module].sort()
        moduleConnections.add(`${sourceModule}-${targetModule}`)
      }
    }

    for (const secret of module.secrets || []) {
      // For each reference in the verb
      for (const ref of secret.references || []) {
        // Only create verb-to-verb child edges
        edges.push(createChildEdge(ref.module, ref.name, module.name, secret.secret?.name || ''))

        // Track module-to-module connection for all reference types
        const [sourceModule, targetModule] = [module.name, ref.module].sort()
        moduleConnections.add(`${sourceModule}-${targetModule}`)
      }
    }

    for (const database of module.databases || []) {
      // For each reference in the verb
      for (const ref of database.references || []) {
        // Only create verb-to-verb child edges
        edges.push(createChildEdge(ref.module, ref.name, module.name, database.database?.name || ''))

        // Track module-to-module connection for all reference types
        const [sourceModule, targetModule] = [module.name, ref.module].sort()
        moduleConnections.add(`${sourceModule}-${targetModule}`)
      }
    }

    for (const topic of module.topics || []) {
      // For each reference in the verb
      for (const ref of topic.references || []) {
        // Only create verb-to-verb child edges
        edges.push(createChildEdge(ref.module, ref.name, module.name, topic.topic?.name || ''))

        // Track module-to-module connection for all reference types
        const [sourceModule, targetModule] = [module.name, ref.module].sort()
        moduleConnections.add(`${sourceModule}-${targetModule}`)
      }
    }
  }

  // Create module-level edges for each unique module connection
  for (const connection of moduleConnections) {
    const [sourceModule, targetModule] = connection.split('-')
    edges.push(createModuleEdge(sourceModule, targetModule))
  }

  return edges
}

export const getGraphData = (
  modules: StreamModulesResult | undefined,
  isDarkMode: boolean,
  nodePositions: Record<string, { x: number; y: number }> = {},
): ElementDefinition[] => {
  if (!modules) return []

  return [
    ...modules.modules.map((module) => createParentNode(module, nodePositions)),
    ...modules.modules.flatMap((module) => createModuleChildren(module, nodePositions, isDarkMode)),
    ...createEdges(modules.modules),
  ]
}

const nodeId = (moduleName: string, name?: string) => {
  if (!name) return moduleName
  return `${moduleName}.${name}`
}
