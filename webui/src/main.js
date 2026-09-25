// Bootstrap is a regular dependency now: import the CSS and only the JS plugin
// actually used by the template (collapse powers the mobile navbar toggler).
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap/js/dist/collapse'

import {createApp, reactive} from 'vue'
import App from './App.vue'
import router from './router'
import axios from './services/axios.js';
import ErrorMsg from './components/ErrorMsg.vue'
import LoadingSpinner from './components/LoadingSpinner.vue'

import './assets/dashboard.css'
import './assets/main.css'

const app = createApp(App)
app.config.globalProperties.$axios = axios;
app.component("ErrorMsg", ErrorMsg);
app.component("LoadingSpinner", LoadingSpinner);
app.use(router)
app.mount('#app')
