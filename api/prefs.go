package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
)

// structs for data
type Pref struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// responses
type PrefResponse struct {
	Status       string `json:"status"`
	ReturnedPref Pref   `json:"pref"`
}

func InitPrefsAPI(e *echo.Echo) {
	e.GET("/prefs/get/:key", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", GetSessionUserID(&c), c.Param("key"))
		if err != nil {
			log.Println("Error while getting pref: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		if !rows.Next() {
			return c.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		}

		resp := Pref{-1, "", ""}
		rows.Scan(&resp.ID, &resp.Key, &resp.Value)

		return c.JSON(http.StatusOK, PrefResponse{"ok", resp})
	})

	e.POST("/prefs/set", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("key") == "" || c.FormValue("value") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", GetSessionUserID(&c), c.FormValue("key"))
		if err != nil {
			log.Println("Error while setting pref: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		if !rows.Next() {
			// doesn't exist, add it
			stmt, err := DB.Prepare("INSERT INTO prefs(userId, `key`, `value`) VALUES(?, ?, ?)")
			if err != nil {
				log.Println("Error while inserting pref: ")
				log.Println(err)
				return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
			}
			_, err = stmt.Exec(GetSessionUserID(&c), c.FormValue("key"), c.FormValue("value"))
			if err != nil {
				log.Println("Error while inserting pref: ")
				log.Println(err)
				return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
			}
			return c.JSON(http.StatusOK, StatusResponse{"ok"})
		} else {
			// exists already, update it
			stmt, err := DB.Prepare("UPDATE prefs SET `key` = ?, `value` = ? WHERE userId = ? AND `key` = ?")
			if err != nil {
				log.Println("Error while updating pref: ")
				log.Println(err)
				return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
			}
			_, err = stmt.Exec(c.FormValue("key"), c.FormValue("value"), GetSessionUserID(&c), c.FormValue("key"))
			if err != nil {
				log.Println("Error while updating pref: ")
				log.Println(err)
				return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
			}
			return c.JSON(http.StatusOK, StatusResponse{"ok"})
		}
	})
}
