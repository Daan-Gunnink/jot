import { createWebHistory, createRouter} from 'vue-router'

import JotView from './views/JotView.vue'

const routes = [
  { path: '/', component: JotView },
  { path: '/jot/:id', component: JotView },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
