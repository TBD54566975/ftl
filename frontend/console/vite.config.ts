import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react()],
  css: {
    modules: {
      localsConvention: 'camelCaseOnly',
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-graph': ['cytoscape'],
          'vendor-fcose': ['cytoscape-fcose'],
          'vendor-ui': ['@headlessui/react', '@tanstack/react-query'],
          'vendor-json-schema-faker': ['json-schema-faker'],
          'vendor-json-schema-library': ['json-schema-library'],
          'vendor-yaml': ['yaml'],
          'vendor-highlightjs': ['highlight.js'],
          'vendor-fuse': ['fuse.js'],
          'vendor-codemirror': [
            '@codemirror/autocomplete',
            '@codemirror/commands',
            '@codemirror/lang-json',
            '@codemirror/language',
            '@codemirror/lint',
            '@codemirror/state',
            '@codemirror/view',
          ],
          'vendor-protos': [
            './src/protos/xyz/block/ftl/console/v1/console_pb',
            './src/protos/xyz/block/ftl/console/v1/console_connect',
            './src/protos/xyz/block/ftl/schema/v1/schema_pb',
            './src/protos/xyz/block/ftl/timeline/v1/event_pb',
            './src/protos/xyz/block/ftl/timeline/v1/timeline_pb',
            './src/protos/xyz/block/ftl/timeline/v1/timeline_connect',
          ],
        },
      },
    },
  },
})
