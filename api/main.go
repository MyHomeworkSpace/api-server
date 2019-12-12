package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/config"
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
	authLevelInternal
)

var DB *sql.DB
var RedisClient *redis.Client

type statusResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
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
		// deal with preflights
		if ec.Request().Method == "OPTIONS" {
			ec.Response().Header().Set("Access-Control-Allow-Credentials", "false")
			ec.Response().Header().Set("Access-Control-Allow-Origin", "*")
			ec.Response().Header().Set("Access-Control-Allow-Headers", "authorization")
			ec.Response().Writer.WriteHeader(http.StatusOK)
			return nil
		}

		// handle cors
		if config.GetCurrent().CORS.Enabled && len(config.GetCurrent().CORS.Origins) > 0 {
			foundOrigin := ""
			for _, origin := range config.GetCurrent().CORS.Origins {
				if origin == ec.Request().Header.Get("Origin") {
					foundOrigin = origin
				}
			}

			if foundOrigin == "" {
				foundOrigin = config.GetCurrent().CORS.Origins[0]
			}

			ec.Response().Header().Set("Access-Control-Allow-Origin", foundOrigin)
			ec.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// some routes bypass session stuff
		bypassSession := strings.HasPrefix(ec.Request().URL.Path, "/application/requestAuth") || strings.HasPrefix(ec.Request().URL.Path, "/auth/completeEmailStart")
		if !bypassSession {
			_, err := ec.Cookie("session")
			if err != nil {
				// user has no cookie, generate one
				cookie := new(http.Cookie)
				cookie.Name = "session"
				cookie.Path = "/"
				uid, err := auth.GenerateUID()
				if err != nil {
					return err
				}
				cookie.Value = uid
				cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
				ec.SetCookie(cookie)
			}

			bypassCSRF := false

			// check if they have an authorization header
			if ec.Request().Header.Get("Authorization") != "" {
				// get the token
				headerParts := strings.Split(ec.Request().Header.Get("Authorization"), " ")
				if len(headerParts) == 2 {
					authToken := headerParts[1]

					// look up token
					rows, err := DB.Query("SELECT applications.cors FROM application_authorizations INNER JOIN applications ON application_authorizations.applicationId = applications.id WHERE application_authorizations.token = ?", authToken)
					if err == nil {
						// IMPORTANT: if there's an error with the token, we just continue with the request
						// this is for backwards compatibility with old versions, where the token would always bypass csrf and only be checked when authentication was needed
						// this is ok because if someone is able to add a new header, it should not be in a scenario where csrf would be a useful defense
						// TODO: it would be much cleaner to just fail here if the token is bad. do any applications actually rely on this behavior?

						defer rows.Close()
						if rows.Next() {
							cors := ""
							err = rows.Scan(&cors)

							if err == nil && cors != "" {
								ec.Response().Header().Set("Access-Control-Allow-Origin", cors)
								ec.Response().Header().Set("Access-Control-Allow-Headers", "authorization")
							}
						}
					}
				}

				// also bypass csrf
				bypassCSRF = true
			}

			// bypass csrf for special internal api (this requires the ip to be localhost so it's still secure)
			if strings.HasPrefix(ec.Request().URL.Path, "/internal") {
				bypassCSRF = true
			}

			if !bypassCSRF {
				csrfCookie, err := ec.Cookie("csrfToken")
				csrfToken := ""
				hasNoToken := false
				if err != nil {
					// user has no cookie, generate one
					cookie := new(http.Cookie)
					cookie.Name = "csrfToken"
					cookie.Path = "/"
					uid, err := auth.GenerateRandomString(40)
					if err != nil {
						return err
					}
					cookie.Value = uid
					cookie.Expires = time.Now().Add(12 * 4 * 7 * 24 * time.Hour)
					ec.SetCookie(cookie)

					hasNoToken = true
					csrfToken = cookie.Value

					// let the next if block handle this
				} else {
					csrfToken = csrfCookie.Value
				}

				// bypass csrf token for /auth/csrf
				if strings.HasPrefix(ec.Request().URL.Path, "/auth/csrf") {
					// did we just make up a token?
					if hasNoToken {
						// if so, return it
						// auth.go won't know the new token yet
						writeJSON(ec.Response(), http.StatusOK, csrfResponse{"ok", csrfToken})
						return nil
					}

					// we didn't, so just pass the request through
					bypassCSRF = true
				}

				if !bypassCSRF && (csrfToken != ec.QueryParam("csrfToken") || hasNoToken) {
					writeJSON(ec.Response(), http.StatusBadRequest, errorResponse{"error", "csrfToken_invalid"})
					return nil
				}
			}
		}

		context := RouteContext{}

		// is this an internal-only thing?
		if level == authLevelInternal {
			// they need to be from a local ip then

			// are they?
			if isInternalRequest(&ec) {
				// yes, bypass other checks
				f(ec.Response(), ec.Request(), ec, context)
				return nil
			}

			// no, bye
			writeJSON(ec.Response(), http.StatusUnauthorized, errorResponse{"error", "forbidden"})
			return nil
		}

		// are they logged in?
		sessionUserID := GetSessionUserID(&ec)

		if sessionUserID != -1 {
			context.LoggedIn = true
			user, err := data.GetUserByID(sessionUserID)
			if err != nil {
				return err
			}
			context.User = &user
		}

		if level != authLevelNone {
			// are they logged in?
			if !context.LoggedIn {
				// no, bye
				writeJSON(ec.Response(), http.StatusUnauthorized, errorResponse{"error", "logged_out"})
				return nil
			}

			if level == authLevelAdmin {
				// are they an admin?
				if context.User.Level == 0 {
					// no, bye
					writeJSON(ec.Response(), http.StatusUnauthorized, errorResponse{"error", "forbidden"})
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

func writeJSON(w http.ResponseWriter, status int, thing interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(thing)
}

// Init will initialize all available API endpoints
func Init(e *echo.Echo) {
	e.GET("/status", route(routeStatus, authLevelNone))

	e.GET("/admin/getAllFeedback", route(routeAdminGetAllFeedback, authLevelAdmin))
	e.GET("/admin/getFeedbackScreenshot/:id", route(routeAdminGetFeedbackScreenshot, authLevelAdmin))
	e.GET("/admin/getUserCount", route(routeAdminGetUserCount, authLevelAdmin))
	e.POST("/admin/sendEmail", route(routeAdminSendEmail, authLevelAdmin))
	e.POST("/admin/triggerError", route(routeAdminTriggerError, authLevelAdmin))

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

	e.POST("/auth/changeEmail", route(routeAuthChangeEmail, authLevelLoggedIn))
	e.POST("/auth/changePassword", route(routeAuthChangePassword, authLevelLoggedIn))
	e.POST("/auth/clearMigrateFlag", route(routeAuthClearMigrateFlag, authLevelLoggedIn))
	e.GET("/auth/completeEmailStart/:token", route(routeAuthCompleteEmailStart, authLevelNone))
	e.POST("/auth/completeEmail", route(routeAuthCompleteEmail, authLevelNone))
	e.POST("/auth/createAccount", route(routeAuthCreateAccount, authLevelNone))
	e.GET("/auth/csrf", route(routeAuthCsrf, authLevelNone))
	e.POST("/auth/login", route(routeAuthLogin, authLevelNone))
	e.GET("/auth/me", route(routeAuthMe, authLevelLoggedIn))
	e.GET("/auth/logout", route(routeAuthLogout, authLevelLoggedIn))
	e.POST("/auth/resetPassword", route(routeAuthResetPassword, authLevelNone))
	e.POST("/auth/resendVerificationEmail", route(routeAuthResendVerificationEmail, authLevelNone))
	e.GET("/auth/session", route(routeAuthSession, authLevelNone))

	e.POST("/auth/2fa/beginEnroll", route(routeAuth2faBeginEnroll, authLevelLoggedIn))
	e.POST("/auth/2fa/completeEnroll", route(routeAuth2faCompleteEnroll, authLevelLoggedIn))
	e.GET("/auth/2fa/status", route(routeAuth2faStatus, authLevelLoggedIn))
	e.POST("/auth/2fa/unenroll", route(routeAuth2faUnenroll, authLevelLoggedIn))

	e.GET("/calendar/getStatus", route(routeCalendarGetStatus, authLevelLoggedIn))
	e.GET("/calendar/getView", route(routeCalendarGetView, authLevelLoggedIn))

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

	e.POST("/internal/startTask", route(routeInternalStartTask, authLevelInternal))

	e.POST("/notifications/add", route(routeNotificationsAdd, authLevelAdmin))
	e.POST("/notifications/delete", route(routeNotificationsDelete, authLevelAdmin))
	e.GET("/notifications/get", route(routeNotificationsGet, authLevelLoggedIn))

	e.GET("/planner/getWeekInfo/:date", route(routePlannerGetWeekInfo, authLevelLoggedIn))

	e.GET("/prefixes/getDefaultList", route(routePrefixesGetDefaultList, authLevelNone))
	e.GET("/prefixes/getList", route(routePrefixesGetList, authLevelLoggedIn))
	e.POST("/prefixes/delete", route(routePrefixesDelete, authLevelLoggedIn))
	e.POST("/prefixes/add", route(routePrefixesAdd, authLevelLoggedIn))

	e.GET("/prefs/get/:key", route(routePrefsGet, authLevelLoggedIn))
	e.GET("/prefs/getAll", route(routePrefsGetAll, authLevelLoggedIn))
	e.POST("/prefs/set", route(routePrefsSet, authLevelLoggedIn))

	e.POST("/schools/enroll", route(routeSchoolsEnroll, authLevelLoggedIn))
	e.GET("/schools/lookup", route(routeSchoolsLookup, authLevelLoggedIn))
	e.POST("/schools/setEnabled", route(routeSchoolsSetEnabled, authLevelLoggedIn))
	e.POST("/schools/unenroll", route(routeSchoolsUnenroll, authLevelLoggedIn))

	e.GET("/schools/settings/get", route(routeSchoolsSettingsGet, authLevelLoggedIn))
	e.POST("/schools/settings/set", route(routeSchoolsSettingsSet, authLevelLoggedIn))
}
