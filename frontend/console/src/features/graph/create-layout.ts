import type { Edge, Node } from 'reactflow'
import type { Module, Topology } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { verbCalls } from '../modules/decls/verb/verb.utils'
import { declIcon } from '../modules/module.utils'
import { declHeight } from './DeclNode'
import { groupPadding } from './GroupNode'

const groupWidth = 200
const ITEM_SPACING = 10

export const layoutNodes = (modules: Module[], topology: Topology | undefined) => {
  const nodes: Node[] = []
  const edges: Edge[] = []

  topology?.levels.reverse().forEach((level, index) => {
    let groupY = 0

    for (const moduleName of level.modules) {
      const module = modules.find((m) => m.name === moduleName)
      if (!module || module.name === 'builtin') {
        continue
      }

      const verbs = module.verbs
      const secrets = module.secrets
      const configs = module.configs
      const databases = module.databases
      const enums = module.enums
      const datas = module.data

      const x = index * 400
      nodes.push({
        id: module.name ?? '',
        position: { x: x, y: groupY },
        data: { title: module.name, item: module },
        type: 'groupNode',
        draggable: false,
        style: {
          width: groupWidth,
          height: moduleHeight(module),
          zIndex: 0,
        },
      })

      let y = 40
      for (const config of configs) {
        if (!config.config) continue
        nodes.push({
          id: `config-${module.name}.${config.config?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: config.config?.name, item: config, icon: declIcon('config', config.config) },
          type: 'declNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: declHeight,
          },
          draggable: false,
          zIndex: 2,
        })
        y += declHeight + ITEM_SPACING
      }

      for (const secret of secrets) {
        if (!secret.secret) continue
        nodes.push({
          id: `secret-${module.name}.${secret.secret?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: secret.secret?.name, item: secret, icon: declIcon('secret', secret.secret) },
          type: 'declNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: declHeight,
          },
          draggable: false,
          zIndex: 2,
        })
        y += declHeight + ITEM_SPACING
      }

      for (const database of databases) {
        if (!database.database) continue
        nodes.push({
          id: `database-${module.name}.${database.database?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: database.database?.name, item: database, icon: declIcon('database', database.database) },
          type: 'declNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: declHeight,
          },
          draggable: false,
          zIndex: 2,
        })
        y += declHeight + ITEM_SPACING
      }

      for (const enumDecl of enums) {
        if (!enumDecl.enum) continue
        nodes.push({
          id: `enum-${module.name}.${enumDecl.enum?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: enumDecl.enum?.name, item: enumDecl, icon: declIcon('enum', enumDecl.enum) },
          type: 'declNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: declHeight,
          },
          draggable: false,
          zIndex: 2,
        })
        y += declHeight + ITEM_SPACING
      }

      for (const data of datas) {
        if (!data.data) continue
        nodes.push({
          id: `data-${module.name}.${data.data?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: data.data?.name, item: data, icon: declIcon('data', data.data) },
          type: 'declNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: declHeight,
          },
          draggable: false,
          zIndex: 2,
        })
        y += declHeight + ITEM_SPACING
      }

      for (const verb of verbs) {
        if (!verb.verb) continue
        const calls = verbCalls(verb)

        nodes.push({
          id: `${module.name}.${verb.verb?.name}`,
          position: { x: 20, y: y },
          connectable: false,
          data: { title: verb.verb?.name, item: verb, icon: declIcon('verb', verb.verb) },
          type: 'declNode',
          parentNode: module.name,
          style: {
            width: groupWidth - 40,
            height: declHeight,
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
                zIndex: 1,
              })
            }
          }
        })

        y += declHeight + ITEM_SPACING
      }

      groupY += y + 40
    }
  })

  return { nodes, edges }
}

const moduleHeight = (module: Module) => {
  let height = groupPadding
  height += (module.configs?.length ?? 0) * (declHeight + ITEM_SPACING)
  height += (module.secrets?.length ?? 0) * (declHeight + ITEM_SPACING)
  height += (module.data?.length ?? 0) * (declHeight + ITEM_SPACING)
  height += (module.databases?.length ?? 0) * (declHeight + ITEM_SPACING)
  height += (module.enums?.length ?? 0) * (declHeight + ITEM_SPACING)
  height += (module.verbs?.length ?? 0) * (declHeight + ITEM_SPACING)
  if (height > groupPadding) {
    height += ITEM_SPACING
  }
  return height
}
