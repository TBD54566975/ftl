import { TabType, timelineTab, TabSearchParams } from '../providers/tabs-provider'
export const invalidTab = ({ id, type }: {id?: string; type?: string}): string | undefined => {
  //P1 no ID or type
  if(!id || !type) {
    return `Required tab field undefined: ${JSON.stringify({ [TabSearchParams.type]:type, [TabSearchParams.id]:id }, null, 2).replace(/":/g, '" :')}`
  }
  //P2 invalid type
  const invalidType = Object.values(TabType).some(v => v === type)
  if(!invalidType) {
    return `Invalid tab type: ${type}`
  }

  //P3 type is timeline but id is wrong
  if(type === TabType.Timeline) {
    if(id !== timelineTab.id) {
      return `invalid timeline id: ${id}`
    }
  }
  //P4 type is verb but invalid type
  if(type === TabType.Verb) {
    const verbIdArray = id.split('.')
    if(type === TabType.Verb && verbIdArray.length !== 2) {
      return `Invalid verb ${id}`
    }
  }
}
