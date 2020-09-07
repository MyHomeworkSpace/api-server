package cornell

// func (s *school) Enroll(tx *sql.Tx, user *data.User, params map[string]interface{}) (map[string]interface{}, error) {
// 	netidRaw, ok := params["netid"]
// 	passwordRaw, ok2 := params["password"]

// 	if !ok || !ok2 {
// 		return nil, data.SchoolError{Code: "missing_params"}
// 	}

// 	netID, ok := netidRaw.(string)
// 	password, ok2 := passwordRaw.(string)

// 	if !ok || !ok2 || netID == "" || password == "" {
// 		return nil, data.SchoolError{Code: "invalid_params"}
// 	}

// 	cookieJar, _ := cookiejar.New(nil)
// 	c := &http.Client{
// 		Jar: cookieJar,
// 	}

// 	term := config.GetCurrent().Cornell.CurrentTerm

// 	resp, err := c.Get("https://classes.cornell.edu/sascuwalogin/login/redirect?redirectUri=https%3A//classes.cornell.edu/scheduler/roster/" + term)
// 	if err != nil {
// 		return nil, data.SchoolError{Code: "couldnt_reach_cornell"}
// 	}

// 	loginPage := resp.Header.Get("location")

// 	fmt.Println(loginPage)
// }
