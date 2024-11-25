import type { EdgeDefinition, ElementDefinition } from 'cytoscape'
import type { StreamModulesResult } from '../../api/modules/use-stream-modules'
import type { Config, Module, Secret, Subscription, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { FTLNode } from './GraphPane'

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
) => ({
  group: 'nodes' as const,
  data: {
    id: childId,
    label: childLabel,
    type: 'node',
    nodeType: childType,
    parent: parentName,
    item,
    backgroundColor: childType === 'verb' ? '#e3f2fd' : childType === 'config' ? '#e8f5e9' : childType === 'secret' ? '#fce4ec' : '#ffffff',
  },
  ...(nodePositions[childId] && {
    position: nodePositions[childId],
  }),
})

const createModuleChildren = (module: Module, nodePositions: Record<string, { x: number; y: number }>) => {
  const children = [
    // Create nodes for verbs
    ...(module.verbs || []).map((verb: Verb) =>
      createChildNode(module.name, `${module.name}-verb-${verb.verb?.name}`, verb.verb?.name || '', 'verb', nodePositions, verb),
    ),
    // Create nodes for configs
    ...(module.configs || []).map((config: Config) =>
      createChildNode(module.name, `${module.name}-config-${config.config?.name}`, config.config?.name || '', 'config', nodePositions, config),
    ),
    // Create nodes for secrets
    ...(module.secrets || []).map((secret: Secret) =>
      createChildNode(module.name, `${module.name}-secret-${secret.secret?.name}`, secret.secret?.name || '', 'secret', nodePositions, secret),
    ),
    // Create nodes for subscriptions
    ...(module.subscriptions || []).map((subscription: Subscription) =>
      createChildNode(
        module.name,
        `${module.name}-subscription-${subscription.subscription?.name}`,
        subscription.subscription?.name || '',
        'subscription',
        nodePositions,
        subscription,
      ),
    ),
  ]
  return children
}

const createChildEdge = (sourceModule: string, sourceVerb: string, targetModule: string, targetVerb: string) => ({
  group: 'edges' as const,
  data: {
    id: `${sourceModule}-${sourceVerb}-to-${targetModule}-${targetVerb}`,
    source: `${sourceModule}-verb-${sourceVerb}`,
    target: `${targetModule}-verb-${targetVerb}`,
    type: 'childConnection',
  },
})

const createModuleEdge = (sourceModule: string, targetModule: string) => ({
  group: 'edges' as const,
  data: {
    id: `module-${sourceModule}-to-${targetModule}`,
    source: sourceModule,
    target: targetModule,
    type: 'moduleConnection',
  },
})

const createVerbEdges = (modules: Module[]) => {
  const edges: EdgeDefinition[] = []
  const moduleConnections = new Set<string>() // Track unique module connections

  for (const module of modules) {
    // For each verb in the module
    for (const verb of module.verbs || []) {
      // For each reference in the verb
      for (const ref of verb.references || []) {
        // Create verb-to-verb edge
        edges.push(createChildEdge(ref.module, ref.name, module.name, verb.verb?.name || ''))

        // Track module-to-module connection
        // Sort module names to ensure consistent edge IDs
        const [sourceModule, targetModule] = [module.name, ref.module].sort()
        moduleConnections.add(`${sourceModule}-${targetModule}`)
      }
    }

    for (const config of module.configs || []) {
      for (const ref of config.references || []) {
        edges.push(createChildEdge(ref.module, ref.name, module.name, config.config?.name || ''))

        // Track module-to-module connection
        // Sort module names to ensure consistent edge IDs
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

export const getGraphData = (modules: StreamModulesResult | undefined, nodePositions: Record<string, { x: number; y: number }> = {}): ElementDefinition[] => {
  if (!modules) return []

  return [
    // Add parent nodes (modules)
    ...modules.modules.map((module) => createParentNode(module, nodePositions)),
    // Add all child nodes
    ...modules.modules.flatMap((module) => createModuleChildren(module, nodePositions)),
    // Add both verb-level and module-level edges
    ...createVerbEdges(modules.modules),
  ]
}
