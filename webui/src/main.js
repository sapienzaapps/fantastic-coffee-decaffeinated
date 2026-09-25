// Bootstrap and Bootstrap Icons are regular dependencies: import their CSS, and
// only the Bootstrap JS plugin actually used by the template (offcanvas powers
// the mobile sidebar).
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap-icons/font/bootstrap-icons.css'
import 'bootstrap/js/dist/offcanvas'

import {createApp} from 'vue'
import App from './App.vue'
import router from './router'
import axios from './services/axios.js';
import ErrorMsg from './components/ErrorMsg.vue'
import LoadingSpinner from './components/LoadingSpinner.vue'

import './assets/theme.css'
import './assets/dashboard.css'

const app = createApp(App)
app.config.globalProperties.$axios = axios;
app.component("ErrorMsg", ErrorMsg);
app.component("LoadingSpinner", LoadingSpinner);
app.use(router)
app.mount('#app')
