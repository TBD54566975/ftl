import { type HugeiconsProps, PackageIcon } from 'hugeicons-react'
import type { StreamModulesResult } from '../../api/modules/use-stream-modules'
import { declIcon, declTypeName, moduleTreeFromStream } from '../modules/module.utils'

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
        url: decl.path,
      })

      if (decl.value && 'fields' in decl.value) {
        for (const field of decl.value.fields ?? []) {
          items.push({
            id: `${module.name}-${decl.value.name}-${field.name}`,
            icon: declIcon(decl.declType, decl.value),
            iconType: declTypeName(decl.declType, decl.value),
            title: field.name,
            subtitle: `${module.name}.${decl.value.name}.${field.name}`,
            url: decl.path,
          })
        }
      }
    }
  }

  return items
}
