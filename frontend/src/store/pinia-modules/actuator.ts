import axios from "axios";
import { ActuatorState, ServerInfo } from "@/types";
import { acceptHMRUpdate, defineStore } from "pinia";

export const useActuatorStore = defineStore("actuator", {
  state: (): ActuatorState => ({
    serverInfo: undefined,
  }),
  getters: {
    info: (state) => {
      return state.serverInfo;
    },
    version: (state) => {
      return state.serverInfo?.version || "";
    },
    isDemo: (state) => {
      return state.serverInfo?.demo || false;
    },
    isReadonly: (state) => {
      return state.serverInfo?.readonly || false;
    },
    needAdminSetup: (state) => {
      return state.serverInfo?.needAdminSetup || false;
    },
  },
  actions: {
    async fetchInfo() {
      const serverInfo = (await axios.get(`/api/actuator/info`))
        .data as ServerInfo;

      this.setInfo(serverInfo);

      return serverInfo;
    },
    setInfo(serverInfo: ServerInfo) {
      this.serverInfo = serverInfo;
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useActuatorStore, import.meta.hot));
}
