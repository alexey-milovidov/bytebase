<template>
  <div class="md:flex md:items-center md:justify-between">
    <div class="flex-1 min-w-0">
      <div class="flex flex-col">
        <div class="flex items-center">
          <div>
            <IssueStatusIcon
              v-if="!create"
              :issue-status="issue.status"
              :task-status="activeTask(issue.pipeline).status"
            />
          </div>
          <BBTextField
            class="ml-2 my-0.5 w-full text-lg font-bold truncate"
            :disabled="!allowEditNameAndDescription"
            :required="true"
            :focus-on-mount="create"
            :bordered="false"
            :value="state.name"
            :placeholder="'Issue name'"
            @end-editing="(text: string) => trySaveName(text)"
          />
        </div>
        <div v-if="!create">
          <i18n-t
            keypath="issue.opened-by-at"
            tag="p"
            class="text-sm text-control-light"
          >
            <template #creator>
              <router-link
                :to="`/u/${issue.creator.id}`"
                class="font-medium text-control hover:underline"
                >{{ issue.creator.name }}</router-link
              >
            </template>
            <template #time>{{
              dayjs(issue.updatedTs * 1000).format("LLL")
            }}</template>
          </i18n-t>
          <p
            v-if="pushEvent"
            class="mt-1 text-sm text-control-light flex flex-row items-center space-x-1"
          >
            <template v-if="pushEvent.vcsType.startsWith('GITLAB')">
              <img class="h-4 w-auto" src="../../assets/gitlab-logo.svg" />
            </template>
            <template v-else-if="pushEvent.vcsType.startsWith('GITHUB')">
              <img class="h-4 w-auto" src="../../assets/github-logo.svg" />
            </template>
            <a :href="vcsBranchUrl" target="_blank" class="normal-link">{{
              `${vcsBranch}@${pushEvent.repositoryFullPath}`
            }}</a>

            <i18n-t
              v-if="commit && commit.id && commit.url"
              keypath="issue.commit-by-at"
              tag="span"
            >
              <template #id>
                <a :href="commit.url" target="_blank" class="normal-link"
                  >{{ commit.id.substring(0, 7) }}:</a
                >
              </template>
              <template #title>
                <span class="text-main">{{ commit.title }}</span>
              </template>
              <template #author>{{ pushEvent.authorName }}</template>
              <template #time>{{
                dayjs(commit.createdTs * 1000).format("LLL")
              }}</template>
            </i18n-t>
          </p>
          <IssueRollbackFromTips />
        </div>
      </div>
    </div>
    <div class="mt-4 flex space-x-3 md:mt-0 md:ml-4">
      <slot />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { reactive, watch, computed, Ref } from "vue";
import IssueStatusIcon from "./IssueStatusIcon.vue";
import IssueRollbackFromTips from "./IssueRollbackFromTips.vue";
import { activeTask } from "@/utils";
import {
  TaskDatabaseSchemaUpdatePayload,
  TaskDatabaseDataUpdatePayload,
  Issue,
  VCSPushEvent,
} from "@/types";
import { useExtraIssueLogic, useIssueLogic } from "./logic";
import { head } from "lodash-es";

interface LocalState {
  editing: boolean;
  name: string;
}

const logic = useIssueLogic();
const create = logic.create;
const issue = logic.issue as Ref<Issue>;
const { allowEditNameAndDescription, updateName } = useExtraIssueLogic();

const state = reactive<LocalState>({
  editing: false,
  name: issue.value.name,
});

watch(
  () => issue.value,
  (curIssue) => {
    state.name = curIssue.name;
  }
);

const pushEvent = computed((): VCSPushEvent | undefined => {
  if (issue.value.type == "bb.issue.database.schema.update") {
    const payload = activeTask(issue.value.pipeline)
      .payload as TaskDatabaseSchemaUpdatePayload;
    return payload?.pushEvent;
  } else if (issue.value.type == "bb.issue.database.data.update") {
    const payload = activeTask(issue.value.pipeline)
      .payload as TaskDatabaseDataUpdatePayload;
    return payload?.pushEvent;
  }
  return undefined;
});

const commit = computed(() => {
  // Use commits[0] for new format
  // Use fileCommit for legacy data (if possible)
  // Use undefined otherwise
  return head(pushEvent.value?.commits) ?? pushEvent.value?.fileCommit;
});

const vcsBranch = computed((): string => {
  if (pushEvent.value) {
    return pushEvent.value.ref.replace(/^refs\/heads\//g, "");
  }
  return "";
});

const vcsBranchUrl = computed((): string => {
  if (pushEvent.value) {
    if (pushEvent.value.vcsType == "GITLAB_SELF_HOST") {
      return `${pushEvent.value.repositoryUrl}/-/tree/${vcsBranch.value}`;
    } else if (pushEvent.value.vcsType == "GITHUB_COM") {
      return `${pushEvent.value.repositoryUrl}/tree/${vcsBranch.value}`;
    }
  }
  return "";
});

const trySaveName = (text: string) => {
  state.name = text;
  if (text != issue.value.name) {
    updateName(state.name);
  }
};
</script>
