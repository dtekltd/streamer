import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { Quasar, Notify, Loading } from 'quasar'
import '@quasar/extras/material-icons/material-icons.css'
import 'quasar/dist/quasar.css'

import App from './App.vue'
import router from './router'
import './styles/app.css'

createApp(App)
  .use(createPinia())
  .use(Quasar, {
    plugins: { Notify, Loading },
    config: {
      iconSet: 'material-icons'
    }
  })
  .use(router)
  .mount('#app')
