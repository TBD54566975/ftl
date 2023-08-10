import { intrinsicElement } from './plop/intrinsic-elements.mjs'
import { voidTags } from './plop/void-tags.mjs'
import Handlerbars from 'handlebars'
// eslint-disable-next-line func-names
export default function (plop) {
  plop.setHelper(
    'getHTMLElement',
    tag => intrinsicElement.get(tag.toLowerCase()) ||
      'HTMLElement'
  )
  plop.setHelper(
    'setRootTag',
    tag => {
      const lowerCased = tag.toLowerCase()
      return new Handlerbars.SafeString(voidTags.has(lowerCased) ? `<${lowerCased} ref={ref} />` : `<${lowerCased} ref={ref}></${lowerCased}>`)
    }
  )
  plop.setGenerator('component', {
    description: 'Create a component',
    prompts: [
      { type: 'input', name: 'name', message: "What is the component's name?" },
      { type: 'input', name: 'tag', message: "What is the component's root tag?" },
    ],
    actions: [
      {
        type: 'addMany',
        destination: 'src/components/{{dashCase name}}',
        base: 'plop/components-templates/',
        templateFiles: 'plop/components-templates/**/*.hbs',
      },
      {
        type: 'append',
        path: 'src/components/index.ts',
        template: `export * from './{{dashCase name}}'`,
      },
    ],
  })

  plop.setGenerator('sub-component', {
    description: 'Create a sub-component',
    prompts: [
      { type: 'input', name: 'parent', message: "What is the parent component's name?" },
      { type: 'input', name: 'name', message: "What is the component's name?" },
      { type: 'input', name: 'tag', message: "What is the component's root tag?" },
    ],
    actions: [
      {
        type: 'add',
        path: 'src/components/{{dashCase parent}}/{{dashCase name}}.tsx',
        templateFile: 'plop/sub-component-templates/component.hbs',
      },
      {
        type: 'add',
        path: 'src/components/{{dashCase parent}}/{{dashCase name}}.spec.tsx',
        templateFile: 'plop/sub-component-templates/spec.tsx.hbs',
      },
    ],
  })
}
