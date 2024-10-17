import { CellsIcon, type HugeiconsProps } from 'hugeicons-react'
import type { PullSchemaResponse } from '../../protos/xyz/block/ftl/v1/ftl_pb'
import { declIcon, declUrl } from '../modules/module.utils'

export interface PaletteItem {
  id: string
  icon: React.FC<Omit<HugeiconsProps, 'ref'> & React.RefAttributes<SVGSVGElement>>
  title: string
  subtitle?: string
  url: string
}

export const paletteItems = (schema: PullSchemaResponse[]): PaletteItem[] => {
  const items: PaletteItem[] = []

  for (const module of schema) {
    items.push({
      id: `${module.moduleName}-module`,
      icon: CellsIcon,
      title: module.moduleName,
      subtitle: module.moduleName,
      url: `/modules/${module.moduleName}`,
    })

    for (const decl of module.schema?.decls ?? []) {
      if (!decl.value || !decl.value.case || !decl.value.value) {
        return []
      }

      items.push({
        id: `${module.moduleName}-${decl.value.value.name}`,
        icon: declIcon(decl.value.case),
        title: decl.value.value.name,
        subtitle: `${module.moduleName}.${decl.value.value.name}`,
        url: declUrl(module.moduleName, decl),
      })

      if (decl.value.case === 'data') {
        for (const field of decl.value.value.fields) {
          items.push({
            id: `${module.moduleName}-${decl.value.value.name}-${field.name}`,
            icon: declIcon(decl.value.case),
            title: field.name,
            subtitle: `${module.moduleName}.${decl.value.value.name}.${field.name}`,
            url: declUrl(module.moduleName, decl),
          })
        }
      }
    }
  }

  return items
}
