import { createApp } from "vue";
import App from "./App.vue";
import "./style.css";

import { createPinia } from "pinia";
import piniaPluginPersistedstate from "pinia-plugin-persistedstate";
import { router } from "./router";
import * as jotService from "./services/jotService";
const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

const app = createApp(App);

app.use(pinia);
app.use(router);
app.mount("#app");

// --- Temporary addition for testing ---
if (import.meta.env.DEV) {
  // Only expose in development mode
  console.log("Exposing jotService test utils to window.jotTestUtils");
  (window as any).jotTestUtils = {
    generate: jotService.generateDummyJots,
    clear: jotService.clearDummyJots,
  };
}
// --- End temporary addition ---
