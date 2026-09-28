package review

import (
	"net/http"
	"strconv"

	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/handlerx"
	"fishing-notes-admin-api/internal/svc"
	"fishing-notes-admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {
	handlerx.AddPermissionRoutes(server, svcCtx, []handlerx.PermissionRoute{
		{PermissionCode: "review.task.read", Route: rest.Route{Method: http.MethodGet, Path: "/review-tasks", Handler: listHandler(svcCtx)}},
		{PermissionCode: "review.task.read", Route: rest.Route{Method: http.MethodGet, Path: "/audit-logs", Handler: auditLogsHandler(svcCtx)}},
		{PermissionCode: "review.task.export", Route: rest.Route{Method: http.MethodGet, Path: "/audit-logs/export", Handler: exportAuditLogsHandler(svcCtx)}},
		{PermissionCode: "public.fishing_report.approve", Route: rest.Route{Method: http.MethodPost, Path: "/review-tasks/public-reports/:recordId/approve", Handler: reviewPublicReportHandler(svcCtx, "approved")}},
		{PermissionCode: "public.fishing_report.reject", Route: rest.Route{Method: http.MethodPost, Path: "/review-tasks/public-reports/:recordId/reject", Handler: reviewPublicReportHandler(svcCtx, "rejected")}},
	})
}

func exportAuditLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAuditLogsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid audit log export query"))
			return
		}
		content, truncated, err := svcCtx.BusinessService.ExportAuditLogs(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="audit-logs.csv"`)
		w.Header().Set("X-Export-Truncated", strconv.FormatBool(truncated))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}

func auditLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAuditLogsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid audit log query"))
			return
		}
		resp, err := svcCtx.BusinessService.ListAuditLogs(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func reviewPublicReportHandler(svcCtx *svc.ServiceContext, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.PublicFishingReportPathRequest
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid public fishing report id"))
			return
		}
		var req types.ReviewPublicFishingReportRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid public fishing report review payload"))
			return
		}
		if err := svcCtx.BusinessService.ReviewPublicFishingReport(r.Context(), pathReq.RecordID, status, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"success": true})
	}
}

func listHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListReviewTasksRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, errorsx.BadRequest("invalid review task query"))
			return
		}
		resp, err := svcCtx.BusinessService.ListReviewTasks(r.Context(), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
