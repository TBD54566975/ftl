import { type HugeiconsProps, PackageIcon } from 'hugeicons-react'
import type { StreamModulesResult } from '../../api/modules/use-stream-modules'
import { Data } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'
import { declIcon, declTypeName, declUrl, moduleTreeFromStream } from '../modules/module.utils'

export interface PaletteItem {
  id: string
  icon: React.FC<Omit<HugeiconsProps, 'ref'> & React.RefAttributes<SVGSVGElement>>
  iconType: string
  title: string
  subtitle?: string
  url: string
}

export const paletteItems = (result: StreamModulesResult): PaletteItem[] => {
  const items: PaletteItem[] = []

  const tree = moduleTreeFromStream(result?.modules || [])

  for (const module of tree) {
    items.push({
      id: `${module.name}-module`,
      icon: PackageIcon,
      iconType: 'module',
      title: module.name,
      subtitle: module.name,
      url: `/modules/${module.name}`,
    })

    for (const decl of module?.decls ?? []) {
      if (!decl.value || !decl.declType || !decl.decl) {
        return []
      }

      items.push({
        id: `${module.name}-${decl.value.name}`,
        icon: declIcon(decl.declType, decl.value),
        iconType: declTypeName(decl.declType, decl.value),
        title: decl.value.name,
        subtitle: `${module.name}.${decl.value.name}`,
        url: declUrl(module.name, decl.decl),
      })

      if (decl.decl instanceof Data) {
        for (const field of decl.decl.fields) {
          items.push({
            id: `${module.name}-${decl.value.name}-${field.name}`,
            icon: declIcon(decl.declType, decl.value),
            iconType: declTypeName(decl.declType, decl.value),
            title: field.name,
            subtitle: `${module.name}.${decl.value.name}.${field.name}`,
            url: declUrl(module.name, decl.decl),
          })
        }
      }
    }
  }

  return items
}
