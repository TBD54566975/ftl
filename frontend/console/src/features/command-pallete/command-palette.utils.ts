import { type HugeiconsProps, PackageIcon } from 'hugeicons-react'
import type { PullSchemaResponse } from '../../protos/xyz/block/ftl/v1/schemaservice_pb'
import { declIcon, declTypeName, declUrl } from '../modules/module.utils'

export interface PaletteItem {
  id: string
  icon: React.FC<Omit<HugeiconsProps, 'ref'> & React.RefAttributes<SVGSVGElement>>
  iconType: string
  title: string
  subtitle?: string
  url: string
}

export const paletteItems = (schema: PullSchemaResponse[]): PaletteItem[] => {
  const items: PaletteItem[] = []

  for (const module of schema) {
    items.push({
      id: `${module.moduleName}-module`,
      icon: PackageIcon,
      iconType: 'module',
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
        icon: declIcon(decl.value.case, decl.value.value),
        iconType: declTypeName(decl.value.case, decl.value.value),
        title: decl.value.value.name,
        subtitle: `${module.moduleName}.${decl.value.value.name}`,
        url: declUrl(module.moduleName, decl),
      })

      if (decl.value.case === 'data') {
        for (const field of decl.value.value.fields) {
          items.push({
            id: `${module.moduleName}-${decl.value.value.name}-${field.name}`,
            icon: declIcon(decl.value.case, decl.value.value),
            iconType: declTypeName(decl.value.case, decl.value.value),
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
