import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/tailwind.css'
import { install as VueMonacoEditorPlugin } from '@guolao/vue-monaco-editor'
import * as monaco from 'monaco-editor'
 
const app = createApp(App)
app.use(router)
app.use(VueMonacoEditorPlugin, {
  monaco
})
app.mount('#app') 