package helper

import (
	"encoding/gob"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	qrterm "github.com/Baozisoftware/qrcode-terminal-go"
	whatsapp "github.com/dimaskiddo/go-whatsapp"
)

var WAConn *whatsapp.Conn

type WAHandler struct {
	SessionConn   *whatsapp.Conn
	SessionJID    string
	SessionFile   string
	SessionStart  uint64
	ReconnectTime int
	IsTest        bool
}

func (wah *WAHandler) HandleError(err error) {
	_, eMatch := err.(*whatsapp.ErrConnectionFailed)
	if eMatch {
		if WASessionExist(wah.SessionFile) && wah.SessionConn != nil {
			log.Println("error connection: connection closed unexpetedly, reconnecting after " + strconv.Itoa(wah.ReconnectTime) + " seconds")

			wah.SessionStart = uint64(time.Now().Unix())
			<-time.After(time.Duration(wah.ReconnectTime) * time.Second)

			err := wah.SessionConn.Restore()
			if err != nil {
				log.Println(strings.ToLower(err.Error()))
			}
		} else {
			log.Println("error connection: connection closed unexpetedly")
		}
	} else {
		if strings.Contains(strings.ToLower(err.Error()), "server closed connection") {
			return
		}

		log.Println(strings.ToLower(err.Error()))
	}
}

func (wah *WAHandler) HandleTextMessage(data whatsapp.TextMessage) {
	msgText := strings.SplitN(strings.ToLower(strings.TrimSpace(data.Text)), " ", 2)
	if msgText[0] != "@bot" || data.Info.Timestamp < wah.SessionStart {
		return
	}

	msgCommand := msgText[1]

	var resText []interface{}
	resText = append(resText, "")

	resCommand, err := CMDExec(CMDList, strings.Split(msgCommand, " "), 0)
	if err == nil {
		resText = SplitAtChar(resCommand.(string), "\n", 2000)
	} else {
		resText[0] = "Ouch, Got some error here while processing your request 🙈"
		log.Println(err.Error())
	}

	if wah.IsTest {
		if data.Info.FromMe && data.Info.RemoteJid == wah.SessionJID {
			for i := 0; i < len(resText); i++ {
				err := WAMessageText(wah.SessionConn, data.Info.RemoteJid, "```"+strings.TrimSpace(resText[i].(string))+"```", 0)
				if err != nil {
					log.Println("handler: error while sending message, " + err.Error())
				}
			}
		}
	} else {
		for i := 0; i < len(resText); i++ {
			err := WAMessageText(wah.SessionConn, data.Info.RemoteJid, "```"+strings.TrimSpace(resText[i].(string))+"```", 0)
			if err != nil {
				log.Println("handler: error while sending message, " + err.Error())
			}
		}
	}
}

func WASessionInit(timeout int) (*whatsapp.Conn, error) {
	conn, err := whatsapp.NewConn(time.Duration(timeout) * time.Second)
	if err != nil {
		return nil, err
	}
	conn.SetClientName("Go WhatsApp CLI", "Go WhatsApp")
	return conn, nil
}

func WASessionLoad(file string) (whatsapp.Session, error) {
	session := whatsapp.Session{}

	buffer, err := os.Open(file)
	if err != nil {
		return session, err
	}
	defer buffer.Close()

	err = gob.NewDecoder(buffer).Decode(&session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func WASessionSave(file string, session whatsapp.Session) error {
	buffer, err := os.Create(file)
	if err != nil {
		return err
	}
	defer buffer.Close()

	err = gob.NewEncoder(buffer).Encode(session)
	if err != nil {
		return err
	}

	return nil
}

func WASessionExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}

	return true
}

func WASessionLogin(conn *whatsapp.Conn, file string) error {
	if conn != nil {
		defer func() {
			_, _ = conn.Disconnect()
		}()

		if WASessionExist(file) {
			return errors.New("error session: session file already exist, please logout first")
		}

		qrstr := make(chan string)
		go func() {
			term := qrterm.New()
			term.Get(<-qrstr).Print()
		}()

		session, err := conn.Login(qrstr)
		if err != nil {
			return err
		}

		err = WASessionSave(file, session)
		if err != nil {
			return err
		}
	} else {
		return errors.New("error connection: connection is not valid")
	}

	return nil
}

func WASessionRestore(conn *whatsapp.Conn, file string) error {
	if conn != nil {
		if !WASessionExist(file) {
			_, _ = conn.Disconnect()
			return errors.New("error session: session file doesn't exist, please login first")
		}

		session, err := WASessionLoad(file)
		if err != nil {
			_ = os.Remove(file)
			_, _ = conn.Disconnect()
			return errors.New("error session: session not valid, removing session file")
		}

		session, err = conn.RestoreWithSession(session)
		if err != nil {
			_, _ = conn.Disconnect()
			return err
		}
	} else {
		return errors.New("error connection: connection is not valid")
	}

	return nil
}

func WASessionLogout(conn *whatsapp.Conn, file string) error {
	if conn != nil {
		defer func() {
			_, _ = conn.Disconnect()
		}()

		err := conn.Logout()
		if err != nil {
			return err
		}

		_ = os.Remove(file)
	} else {
		return errors.New("error connection: connection is not valid")
	}

	return nil
}

func WAMessageText(conn *whatsapp.Conn, msgJID string, msgText string, msgDelay int) error {
	if conn != nil {
		content := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: msgJID,
			},
			Text: msgText,
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		err := conn.Send(content)
		if err != nil {
			return err
		}
	} else {
		return errors.New("error connection: connection is not valid")
	}

	return nil
}
