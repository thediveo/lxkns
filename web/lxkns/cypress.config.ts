import { defineConfig } from "cypress"

export default defineConfig({
  component: {
    devServer: {
      framework: "create-react-app",
      bundler: "webpack",
    },
  },
  env: {
    'cypress-react-selector': {
      root: '#__cy_root',
    },
  },
})
