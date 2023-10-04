import { callIcon, moduleVerbCls, callIconID, vizID } from '../modules.constants'

import styles from './graph.module.css'

export const formatSVG = (svg: SVGSVGElement): SVGSVGElement => {
  svg.insertAdjacentHTML('afterbegin', callIcon)
  svg.removeAttribute('width')
  svg.removeAttribute('height')
  svg.setAttribute('id', vizID)

  for (const $a of svg.querySelectorAll('a')) {
    const $g = $a.parentNode! as SVGSVGElement

    const $docFrag = document.createDocumentFragment()
    while ($a.firstChild) {
      const $child = $a.firstChild
      $docFrag.appendChild($child)
    }

    $g.replaceChild($docFrag, $a)

    $g.id = $g.id.replace(/^a_/, '')
  }
  for (const $el of svg.querySelectorAll('title')) {
    $el.remove()
  }

  for (const $node of svg.querySelectorAll('.node')) {
    $node.classList.remove('node')
    $node.classList.add(styles.node)
  }

  const edgesSources = new Set<string>()
  for (const $edge of svg.querySelectorAll('.edge')) {
    $edge.classList.remove('edge')
    $edge.classList.add(styles.edge)
    const [from, to] = $edge.id.split('=>')
    $edge.removeAttribute('id')
    $edge.setAttribute('data-from', from)
    $edge.setAttribute('data-to', to)
    edgesSources.add(from)
  }

  for (const $el of svg.querySelectorAll('[id*=\\:\\:]')) {
    const [tag, id] = $el.id.split('::')
    if (moduleVerbCls === tag) {
      $el.id = id
    }
    $el.classList.add(styles[tag])
  }

  for (const $path of svg.querySelectorAll(`g.${styles.edge} path`)) {
    const $newPath = $path.cloneNode() as HTMLElement
    $newPath.classList.add(styles.hoverPath)
    $path.classList.add(styles.edgePath)
    $path.parentNode?.appendChild($newPath)
  }
  for (const $verb of svg.querySelectorAll(`.${styles[moduleVerbCls]}`)) {
    const texts = $verb.querySelectorAll('text')
    texts[0].classList.add('verb-name')

    // Tag verb as a call source
    if (edgesSources.has($verb.id)) $verb.classList.add(styles.callSource)

    // Replace icon
    const length = texts.length
    for (let i = 1; i < length; ++i) {
      const str = texts[i].innerHTML
      if (str === '{R}') {
        const $iconPlaceholder = texts[i]
        const height = 22
        const width = 22
        const $useIcon = document.createElementNS('http://www.w3.org/2000/svg', 'use')
        $useIcon.setAttributeNS('http://www.w3.org/1999/xlink', 'href', `#${callIconID}`)
        $useIcon.setAttribute('width', `${width}px`)
        $useIcon.setAttribute('height', `${height}px`)
        $useIcon.classList.add(styles.callLink)
        $useIcon.dataset.verbId = $verb.id

        //hardcoded offset
        const y = parseInt($iconPlaceholder.getAttribute('y')!) - 15
        $useIcon.setAttribute('x', $iconPlaceholder.getAttribute('x')!)
        $useIcon.setAttribute('y', y.toString())
        $verb.replaceChild($useIcon, $iconPlaceholder)
        continue
      }
    }
  }
  return svg
}
