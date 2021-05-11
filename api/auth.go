package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/util"

	"github.com/julienschmidt/httprouter"

	"golang.org/x/crypto/bcrypt"

	"github.com/pquerna/otp/totp"

	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/email"
)

type csrfResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

type contextResponse struct {
	Status string `json:"status"`

	Classes                  []data.HomeworkClass `json:"classes"`
	User                     data.User            `json:"user"`
	Prefixes                 []data.Prefix        `json:"prefixes"`
	PrefixFallbackBackground string               `json:"prefixFallbackBackground"`
	PrefixFallbackColor      string               `json:"prefixFallbackColor"`
	Prefs                    []data.Pref          `json:"prefs"`
	Tabs                     []data.Tab           `json:"tabs"`
}

type meResponse struct {
	Status string `json:"status"`

	User data.User  `json:"user"`
	Tabs []data.Tab `json:"tabs"`

	// for backwards compatibility only
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Level int    `json:"level"`
}

type sessionResponse struct {
	Status  string `json:"status"`
	Session string `json:"session"`
}

type tokenResponse struct {
	Status       string          `json:"status"`
	Token        data.EmailToken `json:"token"`
	InfoRequired bool            `json:"infoRequired"`
}

func sendVerificationEmail(user *data.User) error {
	tokenString, err := util.GenerateRandomString(64)
	if err != nil {
		return err
	}

	err = data.SaveEmailToken(data.EmailToken{
		Token:    tokenString,
		Type:     data.EmailTokenVerifyEmail,
		Metadata: "",
		UserID:   user.ID,
	})
	if err != nil {
		return err
	}

	err = email.Send("", user, "verifyEmail", map[string]interface{}{
		"url": config.GetCurrent().Server.APIURLBase + "auth/completeEmailStart/" + tokenString,
	})
	if err != nil {
		return err
	}

	return nil
}

func HasAuthToken(r *http.Request) bool {
	return r.Header.Get("Authorization") != ""
}

func GetAuthToken(r *http.Request) string {
	headerParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(headerParts) != 2 {
		return ""
	} else {
		return headerParts[1]
	}
}

func GetSessionUserID(r *http.Request) int {
	return GetSessionInfo(r).UserID
}

func GetSessionInfo(r *http.Request) auth.SessionInfo {
	if HasAuthToken(r) {
		// we have an authorization header, use that
		token := GetAuthToken(r)
		if token == "" {
			return auth.SessionInfo{-1}
		}
		return auth.GetSessionFromAuthToken(token)
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		return auth.SessionInfo{-1}
	}
	return auth.GetSession(cookie.Value)
}

func isInternalRequest(r *http.Request) bool {
	remoteAddr := r.RemoteAddr
	if config.GetCurrent().Server.ReverseProxyHeader != "" {
		if r.Header.Get(config.GetCurrent().Server.ReverseProxyHeader) != "" {
			header := strings.Split(r.Header.Get(config.GetCurrent().Server.ReverseProxyHeader), ",")
			remoteAddr = strings.TrimSpace(header[len(header)-1])
		}
	}

	if strings.Split(remoteAddr, ":")[0] == "127.0.0.1" || strings.HasPrefix(remoteAddr, "[::1]") {
		return true
	}

	return false
}

func validatePassword(password string) bool {
	return strings.ContainsAny(strings.ToLower(password), "abcdefghijklmnopqrstuvwxyz") && strings.ContainsAny(password, "0123456789") && len(password) >= 8
}

func handlePasswordChange(user *data.User) error {
	return email.Send("", user, "passwordChange", map[string]interface{}{})
}

/*
 * routes
 */
func routeAuthChangeEmail(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("new") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	new := r.FormValue("new")

	if !util.EmailIsValid(new) {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	emailExists, _, err := data.UserExistsWithEmail(new)
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if emailExists {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "email_exists"})
		return
	}

	tokenString, err := util.GenerateRandomString(64)
	if err != nil {
		errorlog.LogError("changing email", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = data.SaveEmailToken(data.EmailToken{
		Token:    tokenString,
		Type:     data.EmailTokenChangeEmail,
		Metadata: new,
		UserID:   c.User.ID,
	})
	if err != nil {
		errorlog.LogError("changing email", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = email.Send(new, c.User, "emailChange", map[string]interface{}{
		"url": config.GetCurrent().Server.APIURLBase + "auth/completeEmailStart/" + tokenString,
	})
	if err != nil {
		errorlog.LogError("changing email", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAuthChangeName(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("new") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	new := r.FormValue("new")

	_, err := DB.Exec("UPDATE users SET name = ? WHERE id = ?", new, c.User.ID)

	if err != nil {
		errorlog.LogError("changing name", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAuthChangePassword(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("current") == "" || r.FormValue("new") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	if !validatePassword(r.FormValue("new")) {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	current := r.FormValue("current")
	new := r.FormValue("new")

	// first verify if the current password was correct
	err := bcrypt.CompareHashAndPassword([]byte(c.User.PasswordHash), []byte(current))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "password_incorrect"})
		return
	} else if err != nil {
		errorlog.LogError("changing password", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// generate their hash
	hash, err := bcrypt.GenerateFromPassword([]byte(new), bcrypt.DefaultCost)
	if err != nil {
		errorlog.LogError("changing password", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// save their hash
	_, err = DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), c.User.ID)
	if err != nil {
		errorlog.LogError("changing password", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = handlePasswordChange(c.User)
	if err != nil {
		errorlog.LogError("changing password", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAuthClearMigrateFlag(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	_, err := DB.Exec("UPDATE users SET showMigrateMessage = 0 WHERE id = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("clearing migration flag", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAuthCompleteEmailStart(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	http.Redirect(w, r, config.GetCurrent().Server.AppURLBase+"completeEmail:"+p.ByName("token"), http.StatusFound)
}

func routeAuthCompleteEmail(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("token") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	token, err := data.GetEmailToken(r.FormValue("token"))
	if err == data.ErrNotFound {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	} else if err != nil {
		errorlog.LogError("completing email", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if token.Type == data.EmailTokenResetPassword {
		if r.FormValue("password") == "" {
			writeJSON(w, http.StatusOK, tokenResponse{"ok", token, true})
			return
		}

		password := r.FormValue("password")

		if !validatePassword(password) {
			writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
			return
		}

		// generate their hash
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			errorlog.LogError("completing email", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		// save their hash
		_, err = DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), token.UserID)
		if err != nil {
			errorlog.LogError("completing email", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		user, err := data.GetUserByID(token.UserID)
		if err != nil {
			errorlog.LogError("completing email", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		err = handlePasswordChange(&user)
		if err != nil {
			errorlog.LogError("completing email", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		writeJSON(w, http.StatusOK, tokenResponse{"ok", token, false})
	} else if token.Type == data.EmailTokenChangeEmail {
		_, err = DB.Exec("UPDATE users SET email = ?, emailVerified = 1 WHERE id = ?", token.Metadata, token.UserID)
		if err != nil {
			errorlog.LogError("completing email", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		writeJSON(w, http.StatusOK, tokenResponse{"ok", token, false})
	} else if token.Type == data.EmailTokenVerifyEmail {
		_, err = DB.Exec("UPDATE users SET emailVerified = 1 WHERE id = ?", token.UserID)
		if err != nil {
			errorlog.LogError("completing email", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		writeJSON(w, http.StatusOK, tokenResponse{"ok", token, false})
	} else {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	err = data.DeleteEmailToken(token)
	if err != nil {
		errorlog.LogError("completing email", err)
	}
}

func routeAuthContext(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	tabs, err := data.GetTabsByUserID(c.User.ID)
	if err != nil {
		errorlog.LogError("getting user context", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// fetch all classes
	classes, err := data.GetClassesForUser(c.User)
	if err != nil {
		errorlog.LogError("getting user context", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// fetch all prefixes
	prefixes, err := data.GetPrefixesForUser(c.User)
	if err != nil {
		errorlog.LogError("getting user context", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// fetch all prefs
	prefRows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting user context", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer prefRows.Close()

	prefs := []data.Pref{}

	for prefRows.Next() {
		pref := data.Pref{}
		prefRows.Scan(&pref.ID, &pref.Key, &pref.Value)
		prefs = append(prefs, pref)
	}

	writeJSON(w, http.StatusOK, contextResponse{
		Status: "ok",

		Classes:                  classes,
		User:                     *c.User,
		Prefixes:                 prefixes,
		PrefixFallbackBackground: data.FallbackBackground,
		PrefixFallbackColor:      data.FallbackColor,
		Prefs:                    prefs,
		Tabs:                     tabs,
	})
}

func routeAuthCreateAccount(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("name") == "" || r.FormValue("email") == "" || r.FormValue("password") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !util.EmailIsValid(email) {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	if !validatePassword(password) {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check they don't already exist
	emailExists, _, err := data.UserExistsWithEmail(email)
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if emailExists {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "email_exists"})
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	nowUnix := time.Now().Unix()

	// doesn't exist, insert new record
	res, err := tx.Exec(
		"INSERT INTO users(name, username, email, password, type, emailVerified, showMigrateMessage, createdAt, lastLoginAt) VALUES(?, '', ?, ?, 'mhs', 0, 0, ?, ?)",
		name, email, string(passwordHash), nowUnix, nowUnix,
	)
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	userID, err := res.LastInsertId()
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// add default classes
	_, err = tx.Exec(
		"INSERT INTO classes(name, teacher, color, sortIndex, userId) VALUES ('Math', '', 'ff4d40', 0, ?), ('History', '', 'ffa540', 1, ?), ('English', '', '40ff73', 2, ?), ('Language', '', '4071ff', 3, ?), ('Science', '', 'ff4086', 4, ?), ('Other', '', '4d4d4d', 5, ?)",
		int(userID), int(userID), int(userID), int(userID), int(userID), int(userID),
	)
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = tx.Commit()
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// check if there's a school they can enroll in
	var schoolResult *data.SchoolResult
	emailParts := strings.Split(email, "@")
	domain := strings.ToLower(strings.TrimSpace(emailParts[1]))
	school, err := MainRegistry.GetSchoolByEmailDomain(domain)
	if err == data.ErrNotFound {
		schoolResult = nil
	} else if err != nil {
		errorlog.LogError("looking up school by email domain", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	} else {
		schoolResult = &data.SchoolResult{
			SchoolID:    school.ID(),
			DisplayName: school.Name(),
			ShortName:   school.ShortName(),
		}
	}

	// sign them in
	session := auth.SessionInfo{
		UserID: int(userID),
	}
	cookie, _ := r.Cookie("session")
	auth.SetSession(cookie.Value, session)

	// send a verification email
	user, err := data.GetUserByID(int(userID))
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = sendVerificationEmail(&user)
	if err != nil {
		errorlog.LogError("creating account", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, schoolResultResponse{"ok", schoolResult})
}

func routeAuthCsrf(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	cookie, _ := r.Cookie("csrfToken")
	writeJSON(w, http.StatusOK, csrfResponse{"ok", cookie.Value})
}

func routeAuthLogin(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("email") == "" || r.FormValue("password") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// we check if they're already in our db
	userRows, err := DB.Query("SELECT id FROM users WHERE email = ?", email)
	if err != nil {
		errorlog.LogError("getting user information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer userRows.Close()
	if userRows.Next() {
		// email is registered
		// this is the fun part

		userID := -1
		userRows.Scan(&userID)

		user, err := data.GetUserByID(userID)
		if err == data.ErrNotFound {
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "user_record_missing"})
			return
		} else if err != nil {
			errorlog.LogError("getting user information", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		needsConversion := false
		// first we check for the easy path: they have a hash stored with us
		if user.PasswordHash != "" {
			// they do
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err == bcrypt.ErrMismatchedHashAndPassword {
				// bye
				writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "password_incorrect"})
				return
			} else if err != nil {
				errorlog.LogError("user login", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}

			// if we got here, no error -> password correct
		} else {
			// they do not, are they a dalton member? (that is, do they have a username?)
			isDalton := strings.HasSuffix(user.Email, "@dalton.org")
			if isDalton {
				// they are
				// this means we must authenticate with dalton

				daltonUsername := strings.Replace(user.Email, "@dalton.org", "", -1)
				_, resp, _, _, err := auth.DaltonLogin(daltonUsername, password)
				if resp != "" || err != nil {
					writeJSON(w, http.StatusUnauthorized, errorResponse{"error", resp})
					return
				}

				// the sign-in worked
				// flag the account for conversion after passing 2fa
				needsConversion = true
			} else {
				errorlog.LogError("user login", errors.New("user is missing password hash"))
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		}

		// now we check for totp
		enrolled2fa, err := isUser2FAEnrolled(userID)
		if err != nil {
			errorlog.LogError("getting user 2fa enrollment status", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		if enrolled2fa {
			if r.FormValue("code") == "" {
				writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "totp_required"})
				return
			}

			secretRows, err := DB.Query("SELECT totp FROM `2fa` WHERE userId = ?", userID)
			if err != nil {
				errorlog.LogError("getting user 2fa secret", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
			defer secretRows.Close()

			secret := ""

			secretRows.Next()
			secretRows.Scan(&secret)

			if !totp.Validate(r.FormValue("code"), secret) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "bad_totp_code"})
				return
			}
		}

		if needsConversion {
			// if we got here, they signed in with dalton

			// generate their hash
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				errorlog.LogError("converting Dalton user", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}

			// save their hash
			_, err = DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), userID)
			if err != nil {
				errorlog.LogError("converting Dalton user", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		}

		// if we've made it this far, they're signed in
		// update their last login time
		_, err = DB.Exec("UPDATE users SET lastLoginAt = ? WHERE id = ?", time.Now().Unix(), userID)
		if err != nil {
			errorlog.LogError("user login", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		// now set their session cookie
		session := auth.SessionInfo{
			UserID: userID,
		}
		cookie, _ := r.Cookie("session")
		auth.SetSession(cookie.Value, session)

		writeJSON(w, http.StatusOK, statusResponse{"ok"})
	} else {
		// email is not registered, bye
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "no_account"})
	}
}

func routeAuthMe(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	tabs, err := data.GetTabsByUserID(c.User.ID)
	if err != nil {
		errorlog.LogError("getting user information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		Status: "ok",

		User: *c.User,
		Tabs: tabs,

		// these are set for backwards compatibility
		ID:    c.User.ID,
		Name:  c.User.Name,
		Email: c.User.Email,
		Level: c.User.Level,
	})
}

func routeAuthLogout(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	cookie, _ := r.Cookie("session")
	newSession := auth.SessionInfo{-1}
	auth.SetSession(cookie.Value, newSession)
	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAuthResendVerificationEmail(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if c.User.EmailVerified {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "already_verified"})
		return
	}

	err := sendVerificationEmail(c.User)
	if err != nil {
		errorlog.LogError("resending verification email", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAuthResetPassword(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("email") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	emailAddress := r.FormValue("email")

	// we check if they're already in our db
	userRows, err := DB.Query("SELECT id FROM users WHERE email = ?", emailAddress)
	if err != nil {
		errorlog.LogError("password reset", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer userRows.Close()
	if userRows.Next() {
		// email is registered

		userID := -1
		userRows.Scan(&userID)

		user, err := data.GetUserByID(userID)
		if err == data.ErrNotFound {
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "user_record_missing"})
			return
		} else if err != nil {
			errorlog.LogError("password reset", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		tokenString, err := util.GenerateRandomString(64)
		if err != nil {
			errorlog.LogError("password reset", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		err = data.SaveEmailToken(data.EmailToken{
			Token:    tokenString,
			Type:     data.EmailTokenResetPassword,
			Metadata: "",
			UserID:   user.ID,
		})
		if err != nil {
			errorlog.LogError("password reset", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		err = email.Send("", &user, "passwordReset", map[string]interface{}{
			"url": config.GetCurrent().Server.APIURLBase + "auth/completeEmailStart/" + tokenString,
		})
		if err != nil {
			errorlog.LogError("password reset", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		writeJSON(w, http.StatusOK, statusResponse{"ok"})
	} else {
		// email is not registered, bye
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "no_account"})
	}
}

func routeAuthSession(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	cookie, err := r.Cookie("session")
	if err != nil {
		writeJSON(w, http.StatusOK, sessionResponse{"ok", ""})
		return
	}
	writeJSON(w, http.StatusOK, sessionResponse{"ok", cookie.Value})
}
