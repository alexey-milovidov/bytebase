<template>
  <div class="space-y-4 overflow-x-hidden w-208 max-w-full transition-all">
    <!-- Project and base, target database selectors -->
    <div class="mt-2 w-full flex flex-col space-y-4">
      <div class="flex space-x-4">
        <div class="flex w-48 items-center">
          {{ $t("database.sync-schema.select-project") }}
        </div>
        <div class="w-full flex flex-row justify-start items-center px-px">
          <ProjectSelect
            class="!w-48 mr-2 shrink-0"
            :disabled="!allowSelectProject"
            :selected-id="state.projectId"
            @select-project-id="(projectId: ProjectId)=>{
              state.projectId = projectId
            }"
          />
        </div>
      </div>

      <div class="w-full flex space-x-4">
        <div class="flex w-48 items-center">
          {{ $t("database.sync-schema.select-database") }}
        </div>
        <div
          class="w-full flex flex-row justify-start items-center px-px relative"
          :class="isValidId(state.projectId) ? '' : 'opacity-50'"
        >
          <div
            class="absolute top-0 left-0 w-full h-full z-10 cursor-not-allowed"
            :class="isValidId(state.projectId) ? 'hidden' : ''"
          ></div>
          <EnvironmentSelect
            class="!w-48 mr-4 shrink-0"
            name="environment"
            :selected-id="state.targetDatabaseInfo.environmentId"
            :select-default="false"
            @select-environment-id="handleTargetEnvironmentSelect"
          />
          <DatabaseSelect
            :selected-id="(state.targetDatabaseInfo.databaseId as DatabaseId)"
            :mode="'USER'"
            :environment-id="state.targetDatabaseInfo.environmentId"
            :project-id="state.projectId"
            :engine-type-list="engineTypeList"
            :sync-status="'OK'"
            :customize-item="true"
            @select-database-id="handleTargetDatabaseSelect"
          >
            <template #customizeItem="{ database }">
              <div class="flex items-center">
                <InstanceEngineIcon :instance="database.instance" />
                <span class="mx-2">{{ database.name }}</span>
                <span>({{ database.instance.name }})</span>
              </div>
            </template>
          </DatabaseSelect>
        </div>
      </div>
      <hr class="mt-4 w-full" />
      <div class="w-full flex space-x-4">
        <div class="flex w-48 items-center">
          {{ $t("database.sync-schema.from-database") }}
        </div>
        <div
          class="w-full flex flex-row justify-start items-center px-px relative"
          :class="isValidId(state.projectId) ? '' : 'opacity-50'"
        >
          <div
            class="absolute top-0 left-0 w-full h-full z-10 cursor-not-allowed"
            :class="isValidId(state.projectId) ? 'hidden' : ''"
          ></div>
          <EnvironmentSelect
            class="!w-48 mr-4 shrink-0"
            name="environment"
            :selected-id="state.baseSchemaInfo.environmentId"
            :select-default="false"
            @select-environment-id="handleBaseEnvironmentSelect"
          />
          <DatabaseSelect
            :selected-id="(state.baseSchemaInfo.databaseId as DatabaseId)"
            :mode="'USER'"
            :environment-id="state.baseSchemaInfo.environmentId"
            :project-id="state.projectId"
            :engine-type-list="allowedEngineTypeList"
            :sync-status="'OK'"
            :customize-item="true"
            @select-database-id="handleBaseDatabaseSelect"
          >
            <template #customizeItem="{ database }">
              <div class="flex items-center">
                <InstanceEngineIcon :instance="database.instance" />
                <span class="mx-2">{{ database.name }}</span>
                <span>({{ database.instance.name }})</span>
              </div>
            </template>
          </DatabaseSelect>
        </div>
      </div>
      <div class="w-full flex space-x-4">
        <div class="flex w-48 items-center">
          {{ $t("common.schema-version") }}
        </div>
        <div
          class="w-full flex flex-row justify-start items-center px-px relative"
          :class="isValidId(state.projectId) ? '' : 'opacity-50'"
        >
          <BBSelect
            class="!grow"
            :selected-item="state.baseSchemaInfo.migrationHistory"
            :item-list="
                databaseMigrationHistoryList(state.baseSchemaInfo.databaseId as DatabaseId)
              "
            :placeholder="$t('change-history.select')"
            :show-prefix-item="true"
            @select-item="(migrationHistory: MigrationHistory) => handleSchemaVersionSelect(migrationHistory)"
          >
            <template #menuItem="{ item: migrationHistory }">
              <div class="flex justify-between mr-2">
                <!-- To enable the ellipsis function only when 
                overflow happens, we have to set a fixed width. -->
                <NEllipsis style="max-width: 460px">
                  {{ migrationHistory.version }}
                  - {{ migrationHistory.description }}
                </NEllipsis>
                <span class="text-control-light">
                  {{
                    dayjs(migrationHistory.updatedTs * 1000).format(
                      "YYYY-MM-DD HH:mm:ss"
                    )
                  }}
                </span>
              </div>
            </template>
            <template v-if="shouldShowMoreVersionButton" #suffixItem>
              <div
                class="w-full pl-3 leading-8 text-gray-600 cursor-pointer hover:text-accent"
                @click="() => (state.showFeatureModal = true)"
              >
                {{ $t("database.sync-schema.more-version") }}
              </div>
            </template>
          </BBSelect>
        </div>
      </div>
    </div>

    <!-- Schema diff container -->
    <div class="w-full flex flex-col justify-start items-start">
      <div class="w-full flex flex-row justify-start items-center mb-2">
        <span>{{ $t("database.sync-schema.schema-diff") }}</span>
      </div>
      <div v-if="!shouldShowDiff" class="w-full border px-4 py-4">
        <p class="w-full text-left py-2">
          {{ $t("database.sync-schema.select-schema-and-target-database-tip") }}
        </p>
      </div>
      <template v-else>
        <div
          v-if="
            state.baseSchemaInfo.migrationHistory?.schema ===
            state.targetDatabaseInfo.currentSchema
          "
          class="w-full border px-4 py-4"
        >
          <p class="w-full text-left py-2">
            {{ $t("database.sync-schema.no-difference-tip") }}
          </p>
        </div>
        <code-diff
          v-else
          class="code-diff-container w-full h-auto max-h-96 overflow-y-auto"
          :old-string="state.targetDatabaseInfo.currentSchema"
          :new-string="state.baseSchemaInfo.migrationHistory?.schema ?? ''"
          output-format="side-by-side"
          data-label="bb-change-history-code-diff-block"
        />
      </template>
    </div>

    <!-- DDL statement editor container -->
    <div class="w-full flex flex-col justify-start items-start">
      <div class="w-full flex flex-col justify-start mb-2 leading-8">
        <div class="flex flex-row justify-start items-center">
          <span>{{ $t("database.sync-schema.synchronize-statements") }}</span>
          <button
            type="button"
            class="btn-icon ml-2"
            @click.prevent="copyStatement"
          >
            <heroicons-outline:clipboard class="h-5 w-5" />
          </button>
          <span
            v-if="shouldShowDiff && hasDiffBetweenCharactorSets"
            class="text-red-600 ml-3"
          >
            {{ $t("database.sync-schema.character-sets-diff-found") }}
          </span>
        </div>
        <div v-if="shouldShowDiff" class="textinfolabel">
          {{ $t("database.sync-schema.synchronize-statements-description") }}
        </div>
      </div>
      <div
        class="whitespace-pre-wrap w-full overflow-hidden border relative"
        :class="shouldShowDiff ? '' : 'opacity-50'"
      >
        <div
          class="absolute top-0 left-0 w-full h-full z-10 cursor-not-allowed"
          :class="shouldShowDiff ? 'hidden' : ''"
        ></div>
        <MonacoEditor
          ref="editorRef"
          class="w-full h-auto max-h-96"
          data-label="bb-issue-sql-editor"
          :value="state.editStatement"
          :auto-focus="false"
          :dialect="(state.engineType as SQLDialect)"
          @change="onStatementChange"
          @ready="updateEditorHeight"
        />
      </div>
    </div>

    <!-- Buttons group -->
    <div class="pt-2 flex items-center justify-between">
      <span></span>
      <div class="flex items-center justify-end">
        <button
          type="button"
          class="btn-normal py-2 px-4"
          @click.prevent="handleCancelButtonClick"
        >
          {{ $t("common.cancel") }}
        </button>
        <button
          :disabled="!allowCreate"
          class="btn-primary ml-3 inline-flex justify-center py-2 px-4"
          @click="handleConfirmButtonClick"
        >
          {{ $t("database.sync-schema.preview-issue") }}
        </button>
      </div>
    </div>
  </div>

  <FeatureModal
    v-if="state.showFeatureModal"
    feature="bb.feature.sync-schema-all-versions"
    @cancel="state.showFeatureModal = false"
  />
</template>

<script lang="ts" setup>
import axios from "axios";
import dayjs from "dayjs";
import { head, isUndefined } from "lodash-es";
import { NEllipsis } from "naive-ui";
import { computed, reactive, ref, watch } from "vue";
import { useEventListener } from "@vueuse/core";
import { useRouter } from "vue-router";
import { CodeDiff } from "v-code-diff";
import { toClipboard } from "@soerenmartius/vue3-clipboard";
import {
  DatabaseId,
  EngineType,
  EnvironmentId,
  MigrationHistory,
  MigrationType,
  ProjectId,
  SQLDialect,
  UNKNOWN_ID,
} from "@/types";
import {
  hasFeature,
  pushNotification,
  useDatabaseStore,
  useInstanceStore,
} from "@/store";
import { isNullOrUndefined } from "@/plugins/demo/utils";
import EnvironmentSelect from "./EnvironmentSelect.vue";
import DatabaseSelect from "./DatabaseSelect.vue";
import MonacoEditor from "./MonacoEditor/MonacoEditor.vue";
import ProjectSelect from "./ProjectSelect.vue";

type LocalState = {
  projectId?: ProjectId;
  baseSchemaInfo: {
    environmentId?: EnvironmentId;
    databaseId?: DatabaseId;
    migrationHistory?: MigrationHistory;
  };
  targetDatabaseInfo: {
    environmentId?: EnvironmentId;
    databaseId?: DatabaseId;
    currentSchema?: string;
  };
  engineType?: EngineType;
  recommandStatement: string;
  editStatement: string;
  showFeatureModal: boolean;
};

const props = withDefaults(
  defineProps<{
    projectId?: ProjectId;
  }>(),
  {
    projectId: undefined,
  }
);
const emit = defineEmits(["dismiss"]);

const router = useRouter();

const editorRef = ref<InstanceType<typeof MonacoEditor>>();
const instanceStore = useInstanceStore();
const databaseStore = useDatabaseStore();

useEventListener(window, "keydown", (e) => {
  if (e.code === "Escape") {
    emit("dismiss");
  }
});

const state = reactive<LocalState>({
  projectId: props.projectId,
  baseSchemaInfo: {
    databaseId: UNKNOWN_ID,
  },
  targetDatabaseInfo: {
    databaseId: UNKNOWN_ID,
  },
  recommandStatement: "",
  editStatement: "",
  showFeatureModal: false,
});

const isValidId = (id: any) => {
  if (isNullOrUndefined(id) || id === UNKNOWN_ID) {
    return false;
  }
  return true;
};

const allowedEngineTypeList: EngineType[] = ["MYSQL", "POSTGRES"];
const allowedMigrationTypeList: MigrationType[] = [
  "BASELINE",
  "MIGRATE",
  "BRANCH",
];

const hasSyncSchemaFeature = computed(() => {
  return hasFeature("bb.feature.sync-schema-all-versions");
});

const allowSelectProject = computed(() => {
  return props.projectId === undefined;
});

const engineTypeList = computed((): EngineType[] => {
  if (isUndefined(state.engineType)) {
    return allowedEngineTypeList;
  } else {
    return [state.engineType];
  }
});

const shouldShowMoreVersionButton = computed(() => {
  return (
    !hasSyncSchemaFeature.value &&
    databaseMigrationHistoryList(state.baseSchemaInfo.databaseId as DatabaseId)
      .length > 0
  );
});

const shouldShowDiff = computed(() => {
  return (
    isValidId(state.projectId) &&
    !isUndefined(state.engineType) &&
    isValidId(state.baseSchemaInfo.environmentId) &&
    isValidId(state.baseSchemaInfo.databaseId) &&
    !isNullOrUndefined(state.baseSchemaInfo.migrationHistory) &&
    isValidId(state.targetDatabaseInfo.environmentId) &&
    isValidId(state.targetDatabaseInfo.databaseId) &&
    !isNullOrUndefined(state.targetDatabaseInfo.currentSchema)
  );
});

const allowCreate = computed(() => {
  return shouldShowDiff.value && state.editStatement !== "";
});

const hasDiffBetweenCharactorSets = computed(() => {
  if (
    !isValidId(state.baseSchemaInfo.databaseId) ||
    !isValidId(state.targetDatabaseInfo.databaseId)
  ) {
    return false;
  }

  const baseDatabase = databaseStore.getDatabaseById(
    state.baseSchemaInfo.databaseId as number
  );
  const targetDatabase = databaseStore.getDatabaseById(
    state.targetDatabaseInfo.databaseId as number
  );

  return (
    baseDatabase.instance.engine === targetDatabase.instance.engine &&
    baseDatabase.characterSet !== targetDatabase.characterSet
  );
});

const databaseMigrationHistoryList = (databaseId: DatabaseId) => {
  const database = databaseStore.getDatabaseById(databaseId);
  const list = instanceStore
    .getMigrationHistoryListByInstanceIdAndDatabaseName(
      database.instance.id,
      database.name
    )
    .filter((migrationHistory) =>
      allowedMigrationTypeList.includes(migrationHistory.type)
    );

  if (!hasSyncSchemaFeature.value) {
    return list.length > 0 ? [head(list)] : [];
  }
  return list;
};

const handleBaseEnvironmentSelect = async (environmentId: EnvironmentId) => {
  if (environmentId !== state.baseSchemaInfo.environmentId) {
    state.baseSchemaInfo.databaseId = UNKNOWN_ID;
  }
  state.baseSchemaInfo.environmentId = environmentId;
};

const handleTargetEnvironmentSelect = async (environmentId: EnvironmentId) => {
  if (environmentId !== state.targetDatabaseInfo.environmentId) {
    state.targetDatabaseInfo.databaseId = UNKNOWN_ID;
  }
  state.targetDatabaseInfo.environmentId = environmentId;
};

const handleBaseDatabaseSelect = async (databaseId: DatabaseId) => {
  if (isValidId(databaseId)) {
    const database = databaseStore.getDatabaseById(databaseId as DatabaseId);
    if (database) {
      state.baseSchemaInfo.environmentId = database.instance.environment.id;
      state.baseSchemaInfo.databaseId = databaseId;
    }
  }
};

const handleTargetDatabaseSelect = async (databaseId: DatabaseId) => {
  if (isValidId(databaseId)) {
    const database = databaseStore.getDatabaseById(databaseId as DatabaseId);
    if (database) {
      state.targetDatabaseInfo.environmentId = database.instance.environment.id;
      state.targetDatabaseInfo.databaseId = databaseId;
    }
  }
};

const handleSchemaVersionSelect = (migrationHistory: MigrationHistory) => {
  if (!hasSyncSchemaFeature.value) {
    state.showFeatureModal = true;
    return;
  }

  state.baseSchemaInfo.migrationHistory = migrationHistory;
};

const handleCancelButtonClick = () => {
  emit("dismiss");
};

const handleConfirmButtonClick = async () => {
  if (state.editStatement === "") {
    pushNotification({
      module: "bytebase",
      style: "CRITICAL",
      title: "Statements shouldn't be empty",
    });
    return;
  }

  const sourceDatabase = databaseStore.getDatabaseById(
    state.baseSchemaInfo.databaseId as DatabaseId
  );
  const targetDatabase = databaseStore.getDatabaseById(
    state.targetDatabaseInfo.databaseId as DatabaseId
  );
  const issueName = `[${targetDatabase.name}] Sync Schema from ${
    sourceDatabase.name
  }(${state.baseSchemaInfo.migrationHistory?.version || ""}) in ${
    sourceDatabase.instance.environment.name
  } @ ${dayjs().format("MM-DD HH:mm")}`;

  const query: Record<string, any> = {
    template: "bb.issue.database.schema.update",
    name: issueName,
    project: state.projectId,
    databaseList: targetDatabase.id,
    sql: state.editStatement,
  };

  router.push({
    name: "workspace.issue.detail",
    params: {
      issueSlug: "new",
    },
    query,
  });
};

const getSchemaDiff = async (
  engineType: EngineType,
  sourceSchema: string,
  targetSchema: string
) => {
  const { data } = await axios.post("/v1/sql/schema/diff", {
    engineType,
    sourceSchema,
    targetSchema,
  });
  return data;
};

const onStatementChange = (value: string) => {
  state.editStatement = value;
  updateEditorHeight();
};

const updateEditorHeight = () => {
  const contentHeight =
    editorRef.value?.editorInstance?.getContentHeight() as number;
  const actualHeight = contentHeight;
  editorRef.value?.setEditorContentHeight(actualHeight);
};

const copyStatement = () => {
  if (!state.editStatement) {
    return;
  }

  toClipboard(state.editStatement).then(() => {
    pushNotification({
      module: "bytebase",
      style: "INFO",
      title: `Statement copied to clipboard.`,
    });
  });
};

// Pre-fetch all migration history for database with project id.
watch(
  () => [state.projectId],
  async () => {
    if (!isValidId(state.projectId)) {
      return;
    }

    const databaseList = await databaseStore.fetchDatabaseListByProjectId(
      state.projectId as ProjectId
    );
    for (const database of databaseList) {
      if (database.syncStatus === "OK") {
        await instanceStore.fetchMigrationHistory({
          instanceId: database.instance.id,
          databaseName: database.name,
        });
      }
    }
  },
  {
    immediate: true,
  }
);

watch(
  () => [state.baseSchemaInfo.databaseId],
  async () => {
    const databaseId = state.baseSchemaInfo.databaseId;
    if (!isValidId(databaseId)) {
      state.baseSchemaInfo.migrationHistory = undefined;
      return;
    }

    const database = databaseStore.getDatabaseById(databaseId as DatabaseId);
    if (database) {
      const migrationHistoryList = (
        await instanceStore.fetchMigrationHistory({
          instanceId: database.instance.id,
          databaseName: database.name,
        })
      ).filter((migrationHistory) =>
        allowedMigrationTypeList.includes(migrationHistory.type)
      );
      // Default select the first migration history.
      state.baseSchemaInfo.migrationHistory = head(migrationHistoryList);
    } else {
      state.baseSchemaInfo.migrationHistory = undefined;
    }
  }
);

watch(
  () => [state.targetDatabaseInfo.databaseId],
  async () => {
    const databaseId = state.targetDatabaseInfo.databaseId;
    if (!isValidId(databaseId)) {
      state.engineType = undefined;
      return;
    }

    const database = databaseStore.getDatabaseById(databaseId as DatabaseId);
    state.targetDatabaseInfo.currentSchema = undefined;
    if (database) {
      state.engineType = database.instance.engine;
      const currentSchema = await useDatabaseStore().fetchDatabaseSchemaById(
        database.id
      );
      state.targetDatabaseInfo.currentSchema = currentSchema;

      if (isValidId(state.baseSchemaInfo.databaseId)) {
        const baseDatabase = databaseStore.getDatabaseById(
          state.baseSchemaInfo.databaseId as DatabaseId
        );
        if (
          !baseDatabase ||
          state.engineType !== baseDatabase.instance.engine
        ) {
          state.baseSchemaInfo.databaseId = UNKNOWN_ID;
        }
      }
    }
  }
);

watch(
  () => [
    state.baseSchemaInfo.migrationHistory?.schema,
    state.targetDatabaseInfo.currentSchema,
  ],
  async () => {
    if (shouldShowDiff.value) {
      const statement = await getSchemaDiff(
        state.engineType as EngineType,
        /* the current schema of the database to be updated */
        state.targetDatabaseInfo.currentSchema ?? "",
        /* the schema to be updated to */
        state.baseSchemaInfo.migrationHistory?.schema ?? ""
      );
      state.recommandStatement = statement;
      state.editStatement = statement;
    } else {
      state.recommandStatement = "";
      state.editStatement = "";
    }
  }
);
</script>

<style>
.code-diff-container .d2h-file-wrapper {
  @apply mb-0;
}
</style>
