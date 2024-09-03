import type { Edge, Node } from 'reactflow'
import type { Module, Topology } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { verbCalls } from '../verbs/verb.utils'
import { configHeight } from './ConfigNode'
import { groupPadding } from './GroupNode'
import { secretHeight } from './SecretNode'
import { verbHeight } from './VerbNode'

const groupWidth = 200
const ITEM_SPACING = 10

export const layoutNodes = (modules: Module[], topology: Topology | undefined) => {
  const nodes: Node[] = []
  const edges: Edge[] = []

  topology?.levels.reverse().forEach((level, index) => {
    let groupY = 0

    for (const moduleName of level.modules) {
      const module = modules.find((m) => m.name === moduleName)
      if (!module) {
        continue
      }

      const verbs = module.verbs
      const secrets = module.secrets
      const configs = module.configs

      const x = index * 400
      nodes.push({
        id: module.name ?? '',
        position: { x: x, y: groupY },
        data: { title: module.name, item: module },
        type: 'groupNode',
        draggable: true,
        style: {
          width: groupWidth,
          height: moduleHeight(module),
          zIndex: 1,
        },
      })

      let y = 40
      for (const secret of secrets) {
        nodes.push({
          id: `secret-${module.name}.${secret.secret?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: secret.secret?.name, item: secret },
          type: 'secretNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: secretHeight,
          },
          draggable: false,
          zIndex: 2,
        })
        y += secretHeight + ITEM_SPACING
      }

      for (const config of configs) {
        nodes.push({
          id: `config-${module.name}.${config.config?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: config.config?.name, item: config },
          type: 'configNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: configHeight,
          },
          draggable: false,
          zIndex: 2,
        })
        y += configHeight + ITEM_SPACING
      }

      for (const verb of verbs) {
        const calls = verbCalls(verb)

        nodes.push({
          id: `${module.name}.${verb.verb?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: verb.verb?.name, item: verb },
          type: 'verbNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: verbHeight,
          },
          draggable: false,
          zIndex: 2,
        })

        const uniqueEdgeIds = new Set<string>()
        calls?.map((metaCall) => {
          for (const call of metaCall.calls) {
            const edgeId = `${module.name}.${verb.verb?.name}-${call.module}.${call.name}`
            if (!uniqueEdgeIds.has(edgeId)) {
              uniqueEdgeIds.add(edgeId)
              edges.push({
                id: edgeId,
                source: `${module.name}.${verb.verb?.name}`,
                target: `${call.module}.${call.name}`,
                style: { stroke: 'rgb(251 113 133)' },
                animated: true,
              })
            }
          }
        })

        y += verbHeight + ITEM_SPACING
      }

      groupY += y + 40
    }
  })

  return { nodes, edges }
}

const moduleHeight = (module: Module) => {
  let height = groupPadding
  height += (module.secrets?.length ?? 0) * (secretHeight + ITEM_SPACING)
  height += (module.configs?.length ?? 0) * (configHeight + ITEM_SPACING)
  height += (module.verbs?.length ?? 0) * (verbHeight + ITEM_SPACING)
  if (height > groupPadding) {
    height += ITEM_SPACING
  }
  return height
}
