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
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/julienschmidt/httprouter"

	"gopkg.in/redis.v5"
)

type routeFunc func(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext)
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

func route(f routeFunc, level authLevel) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// deal with preflights
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "false")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		// handle cors
		if config.GetCurrent().CORS.Enabled && len(config.GetCurrent().CORS.Origins) > 0 {
			foundOrigin := ""
			for _, origin := range config.GetCurrent().CORS.Origins {
				if origin == r.Header.Get("Origin") {
					foundOrigin = origin
				}
			}

			if foundOrigin == "" {
				foundOrigin = config.GetCurrent().CORS.Origins[0]
			}

			w.Header().Set("Access-Control-Allow-Origin", foundOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// some routes bypass session stuff
		bypassSession := strings.HasPrefix(r.URL.Path, "/application/requestAuth") || strings.HasPrefix(r.URL.Path, "/auth/completeEmailStart")
		if !bypassSession {
			_, err := r.Cookie("session")
			if err != nil {
				// user has no cookie, generate one
				cookie := new(http.Cookie)
				cookie.Name = "session"
				cookie.Path = "/"
				uid, err := auth.GenerateUID()
				if err != nil {
					errorlog.LogError("generating random string for session", err)
					writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
					return
				}
				cookie.Value = uid
				cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
				http.SetCookie(w, cookie)
			}

			bypassCSRF := false

			// check if they have an authorization header
			if r.Header.Get("Authorization") != "" {
				// get the token
				headerParts := strings.Split(r.Header.Get("Authorization"), " ")
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
								w.Header().Set("Access-Control-Allow-Origin", cors)
								w.Header().Set("Access-Control-Allow-Headers", "authorization")
							}
						}
					}
				}

				// also bypass csrf
				bypassCSRF = true
			}

			// bypass csrf for special internal api (this requires the ip to be localhost so it's still secure)
			if strings.HasPrefix(r.URL.Path, "/internal") {
				bypassCSRF = true
			}

			if !bypassCSRF {
				csrfCookie, err := r.Cookie("csrfToken")
				csrfToken := ""
				hasNoToken := false
				if err != nil {
					// user has no cookie, generate one
					cookie := new(http.Cookie)
					cookie.Name = "csrfToken"
					cookie.Path = "/"
					uid, err := auth.GenerateRandomString(40)
					if err != nil {
						errorlog.LogError("generating random string", err)
						writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
						return
					}
					cookie.Value = uid
					cookie.Expires = time.Now().Add(12 * 4 * 7 * 24 * time.Hour)
					http.SetCookie(w, cookie)

					hasNoToken = true
					csrfToken = cookie.Value

					// let the next if block handle this
				} else {
					csrfToken = csrfCookie.Value
				}

				// bypass csrf token for /auth/csrf
				if strings.HasPrefix(r.URL.Path, "/auth/csrf") {
					// did we just make up a token?
					if hasNoToken {
						// if so, return it
						// auth.go won't know the new token yet
						writeJSON(w, http.StatusOK, csrfResponse{"ok", csrfToken})
						return
					}

					// we didn't, so just pass the request through
					bypassCSRF = true
				}

				if !bypassCSRF && (csrfToken != r.FormValue("csrfToken") || hasNoToken) {
					writeJSON(w, http.StatusBadRequest, errorResponse{"error", "csrfToken_invalid"})
					return
				}
			}
		}

		context := RouteContext{}

		// is this an internal-only thing?
		if level == authLevelInternal {
			// they need to be from a local ip then

			// are they?
			if isInternalRequest(r) {
				// yes, bypass other checks
				f(w, r, p, context)
				return
			}

			// no, bye
			writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "forbidden"})
			return
		}

		// are they logged in?
		sessionUserID := GetSessionUserID(r)

		if sessionUserID != -1 {
			context.LoggedIn = true
			user, err := data.GetUserByID(sessionUserID)
			if err != nil {
				errorlog.LogError("getting user information for request", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
			context.User = &user
		}

		if level != authLevelNone {
			// are they logged in?
			if !context.LoggedIn {
				// no, bye
				writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "logged_out"})
				return
			}

			if level == authLevelAdmin {
				// are they an admin?
				if context.User.Level == 0 {
					// no, bye
					writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "forbidden"})
					return
				}
			}
		}

		f(w, r, p, context)
	}
}

func routeStatus(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alive"))
}

func writeJSON(w http.ResponseWriter, status int, thing interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(thing)
}

// Init will initialize all available API endpoints
func Init(router *httprouter.Router) {
	router.GET("/status", route(routeStatus, authLevelNone))

	router.GET("/admin/getAllFeedback", route(routeAdminGetAllFeedback, authLevelAdmin))
	router.GET("/admin/getFeedbackScreenshot/:id", route(routeAdminGetFeedbackScreenshot, authLevelAdmin))
	router.GET("/admin/getUserCount", route(routeAdminGetUserCount, authLevelAdmin))
	router.POST("/admin/sendEmail", route(routeAdminSendEmail, authLevelAdmin))
	router.POST("/admin/triggerError", route(routeAdminTriggerError, authLevelAdmin))

	router.POST("/application/completeAuth", route(routeApplicationCompleteAuth, authLevelLoggedIn))
	router.GET("/application/get/:id", route(routeApplicationGet, authLevelLoggedIn))
	router.GET("/application/getAuthorizations", route(routeApplicationGetAuthorizations, authLevelLoggedIn))
	router.GET("/application/requestAuth/:id", route(routeApplicationRequestAuth, authLevelNone))
	router.POST("/application/revokeAuth", route(routeApplicationRevokeAuth, authLevelLoggedIn))
	router.POST("/application/revokeSelf", route(routeApplicationRevokeSelf, authLevelLoggedIn))

	router.POST("/application/manage/create", route(routeApplicationManageCreate, authLevelLoggedIn))
	router.GET("/application/manage/getAll", route(routeApplicationManageGetAll, authLevelLoggedIn))
	router.POST("/application/manage/update", route(routeApplicationManageUpdate, authLevelLoggedIn))
	router.POST("/application/manage/delete", route(routeApplicationManageDelete, authLevelLoggedIn))

	router.POST("/auth/changeEmail", route(routeAuthChangeEmail, authLevelLoggedIn))
	router.POST("/auth/changePassword", route(routeAuthChangePassword, authLevelLoggedIn))
	router.POST("/auth/clearMigrateFlag", route(routeAuthClearMigrateFlag, authLevelLoggedIn))
	router.GET("/auth/completeEmailStart/:token", route(routeAuthCompleteEmailStart, authLevelNone))
	router.POST("/auth/completeEmail", route(routeAuthCompleteEmail, authLevelNone))
	router.POST("/auth/createAccount", route(routeAuthCreateAccount, authLevelNone))
	router.GET("/auth/csrf", route(routeAuthCsrf, authLevelNone))
	router.POST("/auth/login", route(routeAuthLogin, authLevelNone))
	router.GET("/auth/me", route(routeAuthMe, authLevelLoggedIn))
	router.GET("/auth/logout", route(routeAuthLogout, authLevelLoggedIn))
	router.POST("/auth/resetPassword", route(routeAuthResetPassword, authLevelNone))
	router.POST("/auth/resendVerificationEmail", route(routeAuthResendVerificationEmail, authLevelNone))
	router.GET("/auth/session", route(routeAuthSession, authLevelNone))

	router.POST("/auth/2fa/beginEnroll", route(routeAuth2faBeginEnroll, authLevelLoggedIn))
	router.POST("/auth/2fa/completeEnroll", route(routeAuth2faCompleteEnroll, authLevelLoggedIn))
	router.GET("/auth/2fa/status", route(routeAuth2faStatus, authLevelLoggedIn))
	router.POST("/auth/2fa/unenroll", route(routeAuth2faUnenroll, authLevelLoggedIn))

	router.GET("/calendar/getStatus", route(routeCalendarGetStatus, authLevelLoggedIn))
	router.GET("/calendar/getView", route(routeCalendarGetView, authLevelLoggedIn))

	router.GET("/calendar/events/getWeek/:monday", route(routeCalendarEventsGetWeek, authLevelLoggedIn))

	router.POST("/calendar/events/add", route(routeCalendarEventsAdd, authLevelLoggedIn))
	router.POST("/calendar/events/edit", route(routeCalendarEventsEdit, authLevelLoggedIn))
	router.POST("/calendar/events/delete", route(routeCalendarEventsDelete, authLevelLoggedIn))

	router.POST("/calendar/hwEvents/add", route(routeCalendarHWEventsAdd, authLevelLoggedIn))
	router.POST("/calendar/hwEvents/edit", route(routeCalendarHWEventsEdit, authLevelLoggedIn))
	router.POST("/calendar/hwEvents/delete", route(routeCalendarHWEventsDelete, authLevelLoggedIn))

	router.GET("/classes/get", route(routeClassesGet, authLevelLoggedIn))
	router.GET("/classes/get/:id", route(routeClassesGetID, authLevelLoggedIn))
	router.GET("/classes/hwInfo/:id", route(routeClassesHWInfo, authLevelLoggedIn))
	router.POST("/classes/add", route(routeClassesAdd, authLevelLoggedIn))
	router.POST("/classes/edit", route(routeClassesEdit, authLevelLoggedIn))
	router.POST("/classes/delete", route(routeClassesDelete, authLevelLoggedIn))
	router.POST("/classes/swap", route(routeClassesSwap, authLevelLoggedIn))

	router.POST("/feedback/add", route(routeFeedbackAdd, authLevelLoggedIn))

	router.GET("/homework/get", route(routeHomeworkGet, authLevelLoggedIn))
	router.GET("/homework/getForClass/:classId", route(routeHomeworkGetForClass, authLevelLoggedIn))
	router.GET("/homework/getHWView", route(routeHomeworkGetHWView, authLevelLoggedIn))
	router.GET("/homework/getHWViewSorted", route(routeHomeworkGetHWViewSorted, authLevelLoggedIn))
	router.GET("/homework/get/:id", route(routeHomeworkGetID, authLevelLoggedIn))
	router.GET("/homework/getWeek/:monday", route(routeHomeworkGetWeek, authLevelLoggedIn))
	router.GET("/homework/getPickerSuggestions", route(routeHomeworkGetPickerSuggestions, authLevelLoggedIn))
	router.GET("/homework/search", route(routeHomeworkSearch, authLevelLoggedIn))
	router.POST("/homework/add", route(routeHomeworkAdd, authLevelLoggedIn))
	router.POST("/homework/edit", route(routeHomeworkEdit, authLevelLoggedIn))
	router.POST("/homework/delete", route(routeHomeworkDelete, authLevelLoggedIn))
	router.POST("/homework/markOverdueDone", route(routeHomeworkMarkOverdueDone, authLevelLoggedIn))

	router.POST("/internal/startTask", route(routeInternalStartTask, authLevelInternal))

	router.POST("/notifications/add", route(routeNotificationsAdd, authLevelAdmin))
	router.POST("/notifications/delete", route(routeNotificationsDelete, authLevelAdmin))
	router.GET("/notifications/get", route(routeNotificationsGet, authLevelLoggedIn))

	router.GET("/planner/getWeekInfo/:date", route(routePlannerGetWeekInfo, authLevelLoggedIn))

	router.GET("/prefixes/getDefaultList", route(routePrefixesGetDefaultList, authLevelNone))
	router.GET("/prefixes/getList", route(routePrefixesGetList, authLevelLoggedIn))
	router.POST("/prefixes/delete", route(routePrefixesDelete, authLevelLoggedIn))
	router.POST("/prefixes/add", route(routePrefixesAdd, authLevelLoggedIn))

	router.GET("/prefs/get/:key", route(routePrefsGet, authLevelLoggedIn))
	router.GET("/prefs/getAll", route(routePrefsGetAll, authLevelLoggedIn))
	router.POST("/prefs/set", route(routePrefsSet, authLevelLoggedIn))

	router.POST("/schools/enroll", route(routeSchoolsEnroll, authLevelLoggedIn))
	router.GET("/schools/lookup", route(routeSchoolsLookup, authLevelLoggedIn))
	router.POST("/schools/setEnabled", route(routeSchoolsSetEnabled, authLevelLoggedIn))
	router.POST("/schools/unenroll", route(routeSchoolsUnenroll, authLevelLoggedIn))

	router.GET("/schools/settings/get", route(routeSchoolsSettingsGet, authLevelLoggedIn))
	router.POST("/schools/settings/set", route(routeSchoolsSettingsSet, authLevelLoggedIn))
}
