<template>
  <div>
    <BBTable
      :column-list="COLUMN_LIST"
      :data-source="tableRows"
      :show-header="false"
      :left-bordered="true"
      :right-bordered="true"
      :top-bordered="true"
      :bottom-bordered="true"
      :row-clickable="false"
    >
      <template #body="{ rowData: row }: { rowData: TableRow }">
        <BBTableCell :left-padding="4" class="w-[1%]">
          <!-- as narrow as possible -->
          <div
            class="relative w-5 h-5 flex flex-shrink-0 items-center justify-center rounded-full select-none"
            :class="statusIconClass(row.checkResult.status)"
          >
            <template v-if="row.checkResult.status == 'SUCCESS'">
              <heroicons-solid:check class="w-4 h-4" />
            </template>
            <template v-if="row.checkResult.status == 'WARN'">
              <heroicons-outline:exclamation class="h-4 w-4" />
            </template>
            <template v-else-if="row.checkResult.status == 'ERROR'">
              <span class="text-white font-medium text-base" aria-hidden="true"
                >!</span
              >
            </template>
          </div>
        </BBTableCell>
        <BBTableCell
          v-if="showCategoryColumn"
          class="min-w-[4rem] max-w-[6rem] whitespace-nowrap"
        >
          {{ row.category }}
        </BBTableCell>
        <BBTableCell class="w-[12rem] break-all">
          {{ row.title }}
        </BBTableCell>
        <BBTableCell class="w-auto">
          {{ row.checkResult.content }}
          <a
            v-if="row.link"
            class="normal-link"
            :href="row.link.url"
            :target="row.link.target"
          >
            {{ row.link.title }}
          </a>
        </BBTableCell>
      </template>
    </BBTable>
  </div>
</template>

<script lang="ts" setup>
import { computed, PropType } from "vue";
import { useI18n } from "vue-i18n";
import {
  TaskCheckStatus,
  TaskCheckRun,
  TaskCheckResult,
  GeneralErrorCode,
  ruleTemplateMap,
  getRuleLocalization,
  SQLReviewPolicyErrorCode,
  RuleType,
} from "@/types";
import type { BBTableColumn } from "@/bbkit";
import { LocalizedSQLRuleErrorCodes } from "./const";

interface ErrorCodeLink {
  title: string;
  target: string;
  url: string;
}

type TableRow = {
  checkResult: TaskCheckResult;
  category: string;
  title: string;
  link: ErrorCodeLink | undefined;
};

const props = defineProps({
  taskCheckRun: {
    required: true,
    type: Object as PropType<TaskCheckRun>,
  },
});

const { t } = useI18n();

const statusIconClass = (status: TaskCheckStatus) => {
  switch (status) {
    case "SUCCESS":
      return "bg-success text-white";
    case "WARN":
      return "bg-warning text-white";
    case "ERROR":
      return "bg-error text-white";
  }
};

const checkResultList = computed((): TaskCheckResult[] => {
  if (props.taskCheckRun.status == "DONE") {
    return props.taskCheckRun.result.resultList;
  } else if (props.taskCheckRun.status == "FAILED") {
    return [
      {
        status: "ERROR",
        title: t("common.error"),
        code: props.taskCheckRun.code,
        content: props.taskCheckRun.result.detail,
        namespace: "bb.core",
        line: undefined,
      },
    ];
  } else if (props.taskCheckRun.status == "CANCELED") {
    return [
      {
        status: "WARN",
        title: t("common.canceled"),
        code: props.taskCheckRun.code,
        content: "",
        namespace: "bb.core",
        line: undefined,
      },
    ];
  }

  return [];
});

const categoryAndTitle = (checkResult: TaskCheckResult): [string, string] => {
  if (checkResult.code === SQLReviewPolicyErrorCode.EMPTY_POLICY) {
    const title = `${checkResult.title} (${checkResult.code})`;
    return ["", title];
  }
  if (LocalizedSQLRuleErrorCodes.has(checkResult.code)) {
    const rule = ruleTemplateMap.get(checkResult.title as RuleType);
    if (rule) {
      const ruleLocalization = getRuleLocalization(rule.type);
      const key = `sql-review.category.${rule.category.toLowerCase()}`;
      const category = t(key);
      const title = `${ruleLocalization.title} (${checkResult.code})`;
      return [category, title];
    } else {
      return ["", `${checkResult.title} (${checkResult.code})`];
    }
  }

  return ["", checkResult.title];
};

const errorCodeLink = (
  checkResult: TaskCheckResult
): ErrorCodeLink | undefined => {
  switch (checkResult.code) {
    case GeneralErrorCode.OK:
      return;
    case SQLReviewPolicyErrorCode.EMPTY_POLICY:
      return {
        title: t("sql-review.configure-policy"),
        target: "_self",
        url: "/setting/sql-review",
      };
    default: {
      const url = `https://www.bytebase.com/docs/reference/error-code/${
        checkResult.namespace === "bb.advisor" ? "advisor" : "core"
      }?source=console`;
      return {
        title: t("common.view-doc"),
        target: "__blank",
        url: url,
      };
    }
  }
};

const tableRows = computed(() => {
  return checkResultList.value.map<TableRow>((checkResult) => {
    const [category, title] = categoryAndTitle(checkResult);
    const link = errorCodeLink(checkResult);
    return {
      checkResult,
      category,
      title,
      link,
    };
  });
});

const showCategoryColumn = computed((): boolean =>
  tableRows.value.some((row) => row.category !== "")
);

const COLUMN_LIST = computed((): BBTableColumn[] => {
  const STATUS = {
    title: "Status",
  };
  const CATEGORY = {
    title: "Category",
  };
  const TITLE = {
    title: "Title",
  };
  const CONTENT = {
    title: "Detail",
  };
  if (showCategoryColumn.value) {
    return [STATUS, CATEGORY, TITLE, CONTENT];
  }
  return [STATUS, TITLE, CONTENT];
});
</script>
