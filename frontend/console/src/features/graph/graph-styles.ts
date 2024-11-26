import type { Stylesheet } from 'cytoscape'
import colors from 'tailwindcss/colors'

export const createGraphStyles = (isDarkMode: boolean): Stylesheet[] => {
  const theme = {
    primary: isDarkMode ? colors.indigo[400] : colors.indigo[200],
    background: isDarkMode ? colors.slate[700] : colors.slate[200],
    border: isDarkMode ? colors.slate[400] : colors.slate[600],
    arrow: isDarkMode ? colors.slate[300] : colors.slate[700],
    text: isDarkMode ? colors.slate[100] : colors.slate[900],
    selected: {
      bg: isDarkMode ? colors.blue[400] : colors.blue[500],
      border: isDarkMode ? colors.blue[300] : colors.blue[400],
    },
  }

  return [
    {
      selector: 'node',
      style: {
        'background-color': theme.primary,
        label: 'data(label)',
        'text-valign': 'center',
        'text-halign': 'center',
        color: theme.text,
        shape: 'round-rectangle',
        width: '120px',
        height: '40px',
        'text-wrap': 'wrap',
        'text-max-width': '100px',
        'text-overflow-wrap': 'anywhere',
        'font-size': '12px',
        'border-width': '1px',
        'border-color': theme.border,
      },
    },
    {
      selector: 'edge',
      style: {
        width: 2,
        'line-color': theme.arrow,
        'curve-style': 'bezier',
        'target-arrow-shape': 'triangle',
        'target-arrow-color': theme.arrow,
        'arrow-scale': 1,
      },
    },
    {
      selector: '$node > node',
      style: {
        'padding-top': '10px',
        'padding-left': '10px',
        'padding-bottom': '10px',
        'padding-right': '10px',
        'text-valign': 'top',
        'text-halign': 'center',
        'background-color': theme.background,
      },
    },
    {
      selector: 'node[type="groupNode"]',
      style: {
        'background-color': theme.primary,
        shape: 'round-rectangle',
        width: '180px',
        height: '120px',
        'text-valign': 'top',
        'text-halign': 'center',
        'text-wrap': 'wrap',
        'text-max-width': '120px',
        'text-overflow-wrap': 'anywhere',
        'font-size': '14px',
      },
    },
    {
      selector: ':parent',
      style: {
        'text-valign': 'top',
        'text-halign': 'center',
        'background-opacity': 1,
      },
    },
    {
      selector: '.selected',
      style: {
        'background-color': theme.selected.bg,
        'border-width': 2,
        'border-color': theme.selected.border,
      },
    },
    {
      selector: 'node[type="node"]',
      style: {
        'background-color': 'data(backgroundColor)',
        color: theme.text,
        shape: 'round-rectangle',
        width: '100px',
        height: '30px',
        'border-width': '1px',
        'border-color': theme.border,
        'text-wrap': 'wrap',
        'text-max-width': '80px',
        'text-overflow-wrap': 'anywhere',
        'font-size': '11px',
      },
    },
  ]
}

export const nodeColors = {
  light: {
    verb: colors.indigo[300],
    config: colors.sky[200],
    data: colors.gray[200],
    database: colors.blue[200],
    secret: colors.blue[200],
    subscription: colors.violet[200],
    topic: colors.violet[200],
    default: colors.gray[200],
  },
  dark: {
    verb: colors.indigo[600],
    config: colors.blue[500],
    data: colors.gray[700],
    database: colors.blue[600],
    secret: colors.blue[500],
    subscription: colors.violet[600],
    topic: colors.violet[600],
    default: colors.gray[700],
  },
}

export const getNodeBackgroundColor = (isDarkMode: boolean, nodeType: string): string => {
  const theme = isDarkMode ? nodeColors.dark : nodeColors.light
  return theme[nodeType as keyof typeof theme] || theme.default
}
