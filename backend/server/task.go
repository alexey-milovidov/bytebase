package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/jsonapi"
	"github.com/labstack/echo/v4"

	"github.com/bytebase/bytebase/backend/common"
	api "github.com/bytebase/bytebase/backend/legacyapi"
	"github.com/bytebase/bytebase/backend/store"
	"github.com/bytebase/bytebase/backend/utils"
)

func (s *Server) registerTaskRoutes(g *echo.Group) {
	g.PATCH("/pipeline/:pipelineID/task/all", func(c echo.Context) error {
		ctx := c.Request().Context()
		pipelineID, err := strconv.Atoi(c.Param("pipelineID"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Pipeline ID is not a number: %s", c.Param("pipelineID"))).SetInternal(err)
		}

		taskPatch := &api.TaskPatch{
			UpdaterID: c.Get(getPrincipalIDContextKey()).(int),
		}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, taskPatch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformed update task request").SetInternal(err)
		}

		if taskPatch.EarliestAllowedTs != nil && !s.licenseService.IsFeatureEnabled(api.FeatureTaskScheduleTime) {
			return echo.NewHTTPError(http.StatusForbidden, api.FeatureTaskScheduleTime.AccessErrorMessage())
		}

		issue, err := s.store.GetIssueV2(ctx, &store.FindIssueMessage{PipelineID: &pipelineID})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch issue with pipeline ID: %d", pipelineID)).SetInternal(err)
		}
		if issue == nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Issue not found with pipelineID: %d", pipelineID))
		}

		tasks, err := s.store.ListTasks(ctx, &api.TaskFind{PipelineID: &issue.PipelineUID})
		if err != nil {
			return err
		}
		for _, task := range tasks {
			// Skip gh-ost cutover task as this task has no statement.
			if task.Type == api.TaskDatabaseSchemaUpdateGhostCutover {
				continue
			}
			taskPatch := *taskPatch
			taskPatch.ID = task.ID
			// TODO(d): patch tasks in batch.
			if err := s.TaskScheduler.PatchTask(ctx, task, &taskPatch, issue); err != nil {
				return err
			}
		}
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		return nil
	})

	g.PATCH("/pipeline/:pipelineID/task/:taskID", func(c echo.Context) error {
		ctx := c.Request().Context()
		taskID, err := strconv.Atoi(c.Param("taskID"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Task ID is not a number: %s", c.Param("taskID"))).SetInternal(err)
		}

		taskPatch := &api.TaskPatch{
			ID:        taskID,
			UpdaterID: c.Get(getPrincipalIDContextKey()).(int),
		}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, taskPatch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformed update task request").SetInternal(err)
		}

		if taskPatch.EarliestAllowedTs != nil && !s.licenseService.IsFeatureEnabled(api.FeatureTaskScheduleTime) {
			return echo.NewHTTPError(http.StatusForbidden, api.FeatureTaskScheduleTime.AccessErrorMessage())
		}

		task, err := s.store.GetTaskV2ByID(ctx, taskID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update task").SetInternal(err)
		}
		if task == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Task ID not found: %d", taskID))
		}

		issue, err := s.store.GetIssueV2(ctx, &store.FindIssueMessage{PipelineID: &task.PipelineID})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch issue with pipeline ID: %d", task.PipelineID)).SetInternal(err)
		}
		if issue == nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Issue not found with pipelineID: %d", task.PipelineID))
		}

		if taskPatch.Statement != nil {
			// Tenant mode project don't allow updating SQL statement for a single task.
			if issue.Project.TenantMode == api.TenantModeTenant && (task.Type == api.TaskDatabaseSchemaUpdate || task.Type == api.TaskDatabaseSchemaUpdateSDL) {
				return echo.NewHTTPError(http.StatusBadRequest, "cannot update SQL statement of a single task for projects in tenant mode")
			}
		}

		if err := s.TaskScheduler.PatchTask(ctx, task, taskPatch, issue); err != nil {
			return err
		}
		composedTaskPatched, err := s.store.GetTaskByID(ctx, task.ID)
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, composedTaskPatched); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal update task \"%v\" status response", task.Name)).SetInternal(err)
		}
		return nil
	})

	g.PATCH("/pipeline/:pipelineID/task/:taskID/status", func(c echo.Context) error {
		ctx := c.Request().Context()
		taskID, err := strconv.Atoi(c.Param("taskID"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Task ID is not a number: %s", c.Param("taskID"))).SetInternal(err)
		}

		currentPrincipalID := c.Get(getPrincipalIDContextKey()).(int)
		taskStatusPatch := &api.TaskStatusPatch{
			ID:        taskID,
			UpdaterID: currentPrincipalID,
		}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, taskStatusPatch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformed update task status request").SetInternal(err)
		}

		task, err := s.store.GetTaskV2ByID(ctx, taskID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update task status").SetInternal(err)
		}
		if task == nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Task not found with ID %d", taskID))
		}

		ok, err := s.TaskScheduler.CanPrincipalChangeTaskStatus(ctx, currentPrincipalID, task, taskStatusPatch.Status)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate if the principal can change task status").SetInternal(err)
		}
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not allowed to change task status")
		}

		if taskStatusPatch.Status == api.TaskPending {
			instance, err := s.store.GetInstanceV2(ctx, &store.FindInstanceMessage{UID: &task.InstanceID})
			if err != nil {
				return err
			}
			taskCheckRuns, err := s.store.ListTaskCheckRuns(ctx, &store.TaskCheckRunFind{TaskID: &task.ID})
			if err != nil {
				return err
			}
			ok, err = utils.PassAllCheck(task, api.TaskCheckStatusWarn, taskCheckRuns, instance.Engine)
			if err != nil {
				return err
			}
			if !ok {
				return echo.NewHTTPError(http.StatusBadRequest, "The task has not passed all the checks yet")
			}
			stages, err := s.store.ListStageV2(ctx, task.PipelineID)
			if err != nil {
				return err
			}
			activeStage := utils.GetActiveStage(stages)
			if activeStage == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "All tasks are done already")
			}
			if task.StageID != activeStage.ID {
				return echo.NewHTTPError(http.StatusBadRequest, "Tasks in the prior stage are not done yet")
			}
		}

		if taskStatusPatch.Status == api.TaskDone {
			// the user marks the task as DONE, set Skipped to true and SkippedReason to Comment.
			skipped := true
			taskStatusPatch.Skipped = &skipped
			taskStatusPatch.SkippedReason = taskStatusPatch.Comment
		}

		if err := s.TaskScheduler.PatchTaskStatus(ctx, task, taskStatusPatch); err != nil {
			if common.ErrorCode(err) == common.Invalid {
				return echo.NewHTTPError(http.StatusBadRequest, common.ErrorMessage(err))
			}
			if common.ErrorCode(err) == common.NotImplemented {
				return echo.NewHTTPError(http.StatusNotImplemented, common.ErrorMessage(err))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to update task \"%v\" status", task.Name)).SetInternal(err)
		}
		composedTask, err := s.store.GetTaskByID(ctx, task.ID)
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, composedTask); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal update task \"%v\" status response", task.Name)).SetInternal(err)
		}
		return nil
	})

	g.POST("/pipeline/:pipelineID/task/:taskID/check", func(c echo.Context) error {
		ctx := c.Request().Context()
		taskID, err := strconv.Atoi(c.Param("taskID"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Task ID is not a number: %s", c.Param("taskID"))).SetInternal(err)
		}

		task, err := s.store.GetTaskV2ByID(ctx, taskID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update task status").SetInternal(err)
		}
		if task == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Task not found with ID %d", taskID))
		}
		issue, err := s.store.GetIssueV2(ctx, &store.FindIssueMessage{PipelineID: &task.PipelineID})
		if err != nil {
			return err
		}
		project := issue.Project

		if err := s.TaskCheckScheduler.ScheduleCheck(ctx, project, task, c.Get(getPrincipalIDContextKey()).(int)); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to run task check \"%v\"", task.Name)).SetInternal(err)
		}
		composedTask, err := s.store.GetTaskByID(ctx, task.ID)
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, composedTask); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal update task \"%v\" status response", task.Name)).SetInternal(err)
		}
		return nil
	})
}
