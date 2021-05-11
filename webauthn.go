package main

import (
	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/duo-labs/webauthn/webauthn"
)

var WebAuthnHandler *webauthn.WebAuthn

func initWebAuthn() {
	var err error

	WebAuthnHandler, err = webauthn.New(&webauthn.Config{
		RPDisplayName: config.GetCurrent().Webauthn.DisplayName, // Display Name for your site
		RPID:          config.GetCurrent().Webauthn.RPID,        // Generally the FQDN for your site
		RPOrigin:      config.GetCurrent().Webauthn.RPOrigin,    // The origin URL for WebAuthn requests
		RPIcon:        config.GetCurrent().Webauthn.RPIcon,      // Optional icon URL for your site
	})

	if err != nil {
		panic(err)
	}
}
