import * as React from 'react'
import './expect-no-interdependencies.js'

global.React = React

global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
