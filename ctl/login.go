package ctl

import (
	"github.com/spf13/cobra"

	"github.com/dimaskiddo/go-whatsapp-cli/hlp"
	"github.com/dimaskiddo/go-whatsapp-cli/hlp/libs"
)

// Login Variable Structure
var Login = &cobra.Command{
	Use:   "login",
	Short: "Login to WhatsApp Web",
	Long:  "Login to WhatsApp Web",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		timeout, err := hlp.GetEnvInt("WHATSAPP_TIMEOUT")
		if err != nil {
			timeout, err = cmd.Flags().GetInt("timeout")
			if err != nil {
				hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
			}
		}

		file := "./share/session.gob"

		conn, info, err := libs.WASessionInit(timeout)
		if err != nil {
			hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
		}
		hlp.LogPrintln(hlp.LogLevelInfo, info)

		err = libs.WASessionLogin(conn, file)
		if err != nil {
			hlp.LogPrintln(hlp.LogLevelFatal, err.Error())
		}

		hlp.LogPrintln(hlp.LogLevelInfo, "successfully login to whatsapp web")
	},
}

func init() {
	Login.Flags().Int("timeout", 5, "Timeout connection in second(s). Can be override using WHATSAPP_TIMEOUT environment variable")
}
