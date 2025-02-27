<template>
  <div class="relative">
    <button
      id="user-menu"
      type="button"
      class="flex text-sm text-white focus:outline-none focus:shadow-solid"
      aria-label="User menu"
      aria-haspopup="true"
      @click.prevent="$refs.menu.toggle($event)"
      @contextmenu.capture.prevent="$refs.menu.toggle($event)"
    >
      <PrincipalAvatar :principal="currentUser" />
    </button>
    <BBContextMenu ref="menu" class="origin-top-left mt-2 w-48">
      <router-link
        class="px-4 py-3 menu-item"
        :to="`/u/${currentUser.id}`"
        role="menuitem"
      >
        <p class="text-sm flex justify-between">
          <span class="text-main font-medium truncate">
            {{ currentUser.name }}
          </span>
          <span class="text-control">
            {{ $t(`common.role.${currentUser.role.toLowerCase()}`) }}
          </span>
        </p>
        <p class="text-sm text-control truncate">
          {{ currentUser.email }}
        </p>
      </router-link>
      <div class="border-t border-gray-100"></div>
      <div
        v-if="!isRelease"
        class="py-1 menu-item"
        role="menuitem"
        @click.prevent="ping"
      >
        Ping
      </div>
      <div class="border-t border-gray-100"></div>
      <div class="py-1">
        <router-link to="/setting" class="menu-item" role="menuitem">{{
          $t("common.settings")
        }}</router-link>
        <div
          class="menu-item relative"
          @mouseenter="$refs.languageMenu.toggle($event)"
          @mouseleave="$refs.languageMenu.toggle($event)"
          @click.capture.prevent="$refs.languageMenu.toggle($event)"
        >
          <span>{{ $t("common.language") }}</span>
          <BBContextMenu
            ref="languageMenu"
            class="origin-left absolute left-0 top-0 -translate-x-48 transform"
          >
            <div
              class="menu-item px-3 py-1 hover:bg-gray-100"
              :class="{ 'bg-gray-100': locale === 'en-US' }"
              @click.prevent="toggleLocale('en-US')"
            >
              <div class="radio text-sm">
                <input type="radio" class="btn" :checked="locale === 'en-US'" />
                <label class="ml-2">English</label>
              </div>
            </div>
            <div
              class="menu-item px-3 py-1 hover:bg-gray-100"
              :class="{ 'bg-gray-100': locale === 'zh-CN' }"
              @click.prevent="toggleLocale('zh-CN')"
            >
              <div class="radio text-sm">
                <input type="radio" class="btn" :checked="locale === 'zh-CN'" />
                <label class="ml-2">简体中文</label>
              </div>
            </div>
          </BBContextMenu>
        </div>
        <a
          v-if="showQuickstart"
          class="menu-item"
          role="menuitem"
          @click.prevent="resetQuickstart"
          >{{ $t("common.quickstart") }}</a
        >
        <a
          href="https://bytebase.com/docs?source=console"
          target="_blank"
          class="menu-item"
        >
          {{ $t("common.help") }}
        </a>
      </div>
      <div class="border-t border-gray-100"></div>
      <div v-if="allowToggleDebug" class="py-1 menu-item">
        <div class="flex flex-row items-center space-x-2 justify-between">
          <span> Debug </span>
          <BBSwitch :value="isDebug" @toggle="switchDebug" />
        </div>
      </div>
      <div class="border-t border-gray-100"></div>
      <div class="py-1">
        <a class="menu-item" role="menuitem" @click.prevent="logout">{{
          $t("common.logout")
        }}</a>
      </div>
    </BBContextMenu>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, ref } from "vue";
import { useRouter } from "vue-router";
import { storeToRefs } from "pinia";
import { ServerInfo } from "@/types";
import { hasWorkspacePermission } from "@/utils";
import { useLanguage } from "@/composables/useLanguage";
import {
  pushNotification,
  useActuatorStore,
  useAuthStore,
  useDebugStore,
  useUIStateStore,
} from "@/store";
import PrincipalAvatar from "./PrincipalAvatar.vue";

export default defineComponent({
  name: "ProfileDropdown",
  components: { PrincipalAvatar },
  setup() {
    const actuatorStore = useActuatorStore();
    const authStore = useAuthStore();
    const debugStore = useDebugStore();
    const uiStateStore = useUIStateStore();
    const router = useRouter();
    const { setLocale, locale } = useLanguage();
    const languageMenu = ref();

    const { currentUser } = storeToRefs(authStore);

    // For now, debug mode is a global setting and will affect all users.
    // So we only allow DBA and Owner to toggle it.
    const allowToggleDebug = computed(() => {
      return hasWorkspacePermission(
        "bb.permission.workspace.debug",
        currentUser.value.role
      );
    });

    const showQuickstart = computed(() => !actuatorStore.isDemo);

    const logout = () => {
      authStore.logout().then(() => {
        router.push({ name: "auth.signin" });
      });
    };

    const resetQuickstart = () => {
      const keys = [
        "issue.visit",
        "project.visit",
        "environment.visit",
        "instance.visit",
        "database.visit",
        "member.addOrInvite",
        "hidden",
        "guide.project",
        "guide.environment",
        "guide.instance",
        "guide.database",
        "kbar.open",
      ];
      keys.forEach((key) => {
        uiStateStore.saveIntroStateByKey({
          key,
          newState: false,
        });
      });
    };

    const { isDebug } = storeToRefs(debugStore);

    const switchDebug = () => {
      debugStore.patchDebug({
        isDebug: !isDebug.value,
      });
    };

    const ping = () => {
      actuatorStore.fetchServerInfo().then((info: ServerInfo) => {
        pushNotification({
          module: "bytebase",
          style: "SUCCESS",
          title: JSON.stringify(info, null, 4),
        });
      });
    };

    const toggleLocale = (lang: string) => {
      setLocale(lang);
      languageMenu.value.toggle();
    };

    return {
      currentUser,
      allowToggleDebug,
      showQuickstart,
      resetQuickstart,
      logout,
      isDebug,
      switchDebug,
      ping,
      toggleLocale,
      languageMenu,
      locale,
    };
  },
});
</script>
