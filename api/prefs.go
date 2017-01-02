package api

import (
	"log"
	"net/http"

	"gopkg.in/labstack/echo.v2"
)

// structs for data
type Pref struct {
	ID int `json:"id"`
	Key string `json:"key"`
	Value string `json:"value"`
}

// responses
type PrefResponse struct {
	Status string `json:"status"`
	ReturnedPref Pref `json:"pref"`
}

func InitPrefsAPI(e *echo.Echo) {
	e.GET("/prefs/get/:key", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", GetSessionUserID(&c), c.Param("key"))
		if err != nil {
			log.Println("Error while getting pref: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "not_found"}
			return c.JSON(http.StatusNotFound, jsonResp)
		}

		resp := Pref{-1, "", ""}
		rows.Scan(&resp.ID, &resp.Key, &resp.Value)

		jsonResp := PrefResponse{"ok", resp}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/prefs/set", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("key") == "" || c.FormValue("value") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", GetSessionUserID(&c), c.FormValue("key"))
		if err != nil {
			log.Println("Error while setting pref: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		if !rows.Next() {
			// doesn't exist, add it
			stmt, err := DB.Prepare("INSERT INTO prefs(userId, `key`, `value`) VALUES(?, ?, ?)")
			if err != nil {
				log.Println("Error while inserting pref: ")
				log.Println(err)
				jsonResp := StatusResponse{"error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			_, err = stmt.Exec(GetSessionUserID(&c), c.FormValue("key"), c.FormValue("value"))
			if err != nil {
				log.Println("Error while inserting pref: ")
				log.Println(err)
				jsonResp := StatusResponse{"error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			jsonResp := StatusResponse{"ok"}
			return c.JSON(http.StatusOK, jsonResp)
		} else {
			// exists already, update it
			stmt, err := DB.Prepare("UPDATE prefs SET `key` = ?, `value` = ? WHERE userId = ? AND `key` = ?")
			if err != nil {
				log.Println("Error while updating pref: ")
				log.Println(err)
				jsonResp := StatusResponse{"error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			_, err = stmt.Exec(c.FormValue("key"), c.FormValue("value"), GetSessionUserID(&c), c.FormValue("key"))
			if err != nil {
				log.Println("Error while updating pref: ")
				log.Println(err)
				jsonResp := StatusResponse{"error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			jsonResp := StatusResponse{"ok"}
			return c.JSON(http.StatusOK, jsonResp)
		}
	})
}
