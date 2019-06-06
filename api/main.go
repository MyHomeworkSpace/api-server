package api

import (
	"database/sql"
	"net/http"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/labstack/echo"

	"gopkg.in/redis.v5"
)

type routeFunc func(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext)
type authLevel int

const (
	authLevelNone authLevel = iota
	authLevelLoggedIn
	authLevelAdmin
)

var AuthURLBase string
var DB *sql.DB
var RedisClient *redis.Client
var ReverseProxyHeader string

var FeedbackSlackEnabled bool
var FeedbackSlackURL string
var FeedbackSlackHostName string

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// A RouteContext contains information relevant to the current route
type RouteContext struct {
	LoggedIn bool
	User     *data.User
}

func route(f routeFunc, level authLevel) func(ec echo.Context) error {
	return func(ec echo.Context) error {
		context := RouteContext{}

		// are they logged in?
		sessionUserID := GetSessionUserID(&ec)

		if sessionUserID != -1 {
			context.LoggedIn = true
			user, err := Data_GetUserByID(sessionUserID)
			if err != nil {
				return err
			}
			context.User = &user
		}

		if level != authLevelNone {
			// are they logged in?
			if !context.LoggedIn {
				// no, bye
				ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
				return nil
			}

			if level == authLevelAdmin {
				// are they an admin?
				if context.User.Level == 0 {
					// no, bye
					ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
					return nil
				}
			}
		}

		f(ec.Response(), ec.Request(), ec, context)
		return nil
	}
}

func routeStatus(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	ec.String(http.StatusOK, "Alive")
}

// Init will initialize all available API endpoints
func Init(e *echo.Echo) {
	e.GET("/status", route(routeStatus, authLevelNone))

	e.GET("/admin/getAllFeedback", route(routeAdminGetAllFeedback, authLevelAdmin))
	e.GET("/admin/getFeedbackScreenshot/:id", route(routeAdminGetFeedbackScreenshot, authLevelAdmin))
	e.GET("/admin/getUserCount", route(routeAdminGetUserCount, authLevelAdmin))

	e.POST("/application/completeAuth", route(routeApplicationCompleteAuth, authLevelLoggedIn))
	e.GET("/application/get/:id", route(routeApplicationGet, authLevelLoggedIn))
	e.GET("/application/getAuthorizations", route(routeApplicationGetAuthorizations, authLevelLoggedIn))
	e.GET("/application/requestAuth/:id", route(routeApplicationRequestAuth, authLevelNone))
	e.POST("/application/revokeAuth", route(routeApplicationRevokeAuth, authLevelLoggedIn))
	e.POST("/application/revokeSelf", route(routeApplicationRevokeSelf, authLevelLoggedIn))

	e.POST("/application/manage/create", route(routeApplicationManageCreate, authLevelLoggedIn))
	e.GET("/application/manage/getAll", route(routeApplicationManageGetAll, authLevelLoggedIn))
	e.POST("/application/manage/update", route(routeApplicationManageUpdate, authLevelLoggedIn))
	e.POST("/application/manage/delete", route(routeApplicationManageDelete, authLevelLoggedIn))

	e.POST("/auth/clearMigrateFlag", route(routeAuthClearMigrateFlag, authLevelLoggedIn))
	e.GET("/auth/csrf", route(routeAuthCsrf, authLevelNone))
	e.POST("/auth/login", route(routeAuthLogin, authLevelNone))
	e.GET("/auth/me", route(routeAuthMe, authLevelLoggedIn))
	e.GET("/auth/logout", route(routeAuthLogout, authLevelLoggedIn))
	e.GET("/auth/session", route(routeAuthSession, authLevelNone))

	e.POST("/auth/2fa/beginEnroll", route(routeAuth2faBeginEnroll, authLevelLoggedIn))
	e.POST("/auth/2fa/completeEnroll", route(routeAuth2faCompleteEnroll, authLevelLoggedIn))
	e.GET("/auth/2fa/status", route(routeAuth2faStatus, authLevelLoggedIn))
	e.POST("/auth/2fa/unenroll", route(routeAuth2faUnenroll, authLevelLoggedIn))

	e.GET("/calendar/getStatus", route(routeCalendarGetStatus, authLevelLoggedIn))
	e.GET("/calendar/getView", route(routeCalendarGetView, authLevelLoggedIn))
	e.POST("/calendar/import", route(routeCalendarImport, authLevelLoggedIn))
	e.POST("/calendar/resetSchedule", route(routeCalendarResetSchedule, authLevelLoggedIn))

	e.GET("/calendar/events/getWeek/:monday", route(routeCalendarEventsGetWeek, authLevelLoggedIn))

	e.POST("/calendar/events/add", route(routeCalendarEventsAdd, authLevelLoggedIn))
	e.POST("/calendar/events/edit", route(routeCalendarEventsEdit, authLevelLoggedIn))
	e.POST("/calendar/events/delete", route(routeCalendarEventsDelete, authLevelLoggedIn))

	e.POST("/calendar/hwEvents/add", route(routeCalendarHWEventsAdd, authLevelLoggedIn))
	e.POST("/calendar/hwEvents/edit", route(routeCalendarHWEventsEdit, authLevelLoggedIn))
	e.POST("/calendar/hwEvents/delete", route(routeCalendarHWEventsDelete, authLevelLoggedIn))

	e.GET("/classes/get", route(routeClassesGet, authLevelLoggedIn))
	e.GET("/classes/get/:id", route(routeClassesGetID, authLevelLoggedIn))
	e.GET("/classes/hwInfo/:id", route(routeClassesHWInfo, authLevelLoggedIn))
	e.POST("/classes/add", route(routeClassesAdd, authLevelLoggedIn))
	e.POST("/classes/edit", route(routeClassesEdit, authLevelLoggedIn))
	e.POST("/classes/delete", route(routeClassesDelete, authLevelLoggedIn))
	e.POST("/classes/swap", route(routeClassesSwap, authLevelLoggedIn))

	e.POST("/feedback/add", route(routeFeedbackAdd, authLevelLoggedIn))

	e.GET("/homework/get", route(routeHomeworkGet, authLevelLoggedIn))
	e.GET("/homework/getForClass/:classId", route(routeHomeworkGetForClass, authLevelLoggedIn))
	e.GET("/homework/getHWView", route(routeHomeworkGetHWView, authLevelLoggedIn))
	e.GET("/homework/getHWViewSorted", route(routeHomeworkGetHWViewSorted, authLevelLoggedIn))
	e.GET("/homework/get/:id", route(routeHomeworkGetID, authLevelLoggedIn))
	e.GET("/homework/getWeek/:monday", route(routeHomeworkGetWeek, authLevelLoggedIn))
	e.GET("/homework/getPickerSuggestions", route(routeHomeworkGetPickerSuggestions, authLevelLoggedIn))
	e.GET("/homework/search", route(routeHomeworkSearch, authLevelLoggedIn))
	e.POST("/homework/add", route(routeHomeworkAdd, authLevelLoggedIn))
	e.POST("/homework/edit", route(routeHomeworkEdit, authLevelLoggedIn))
	e.POST("/homework/delete", route(routeHomeworkDelete, authLevelLoggedIn))
	e.POST("/homework/markOverdueDone", route(routeHomeworkMarkOverdueDone, authLevelLoggedIn))

	e.POST("/notifications/add", route(routeNotificationsAdd, authLevelLoggedIn))
	e.POST("/notifications/delete", route(routeNotificationsDelete, authLevelLoggedIn))
	e.GET("/notifications/get", route(routeNotificationsGet, authLevelLoggedIn))

	e.GET("/planner/getWeekInfo/:date", route(routePlannerGetWeekInfo, authLevelLoggedIn))

	e.GET("/prefixes/getDefaultList", route(routePrefixesGetDefaultList, authLevelNone))
	e.GET("/prefixes/getList", route(routePrefixesGetList, authLevelLoggedIn))
	e.POST("/prefixes/delete", route(routePrefixesDelete, authLevelLoggedIn))
	e.POST("/prefixes/add", route(routePrefixesAdd, authLevelLoggedIn))

	e.GET("/prefs/get/:key", route(routePrefsGet, authLevelLoggedIn))
	e.GET("/prefs/getAll", route(routePrefsGetAll, authLevelLoggedIn))
	e.POST("/prefs/set", route(routePrefsSet, authLevelLoggedIn))
}
