package email

import (
	"bytes"
	"errors"
	"fmt"
	htmlTpl "html/template"
	"math/rand"
	"mime/quotedprintable"
	"net/smtp"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	textTpl "text/template"
	"time"

	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/data"
)

// ErrDisabled is reported when Send is called while email has been disabled
var ErrDisabled = errors.New("email: sending email has been disabled in the config file")

type templateData struct {
	User    *data.User
	ToEmail string
	Data    map[string]interface{}
}

var hostString string
var auth smtp.Auth
var htmlFuncs htmlTpl.FuncMap
var textFuncs textTpl.FuncMap

func templateGetFirstName(name string) string {
	return strings.Split(name, " ")[0]
}

// Init sets up the email package
func Init() {
	emailConfig := config.GetCurrent().Email

	if !emailConfig.Enabled {
		// kthxbye
		return
	}

	htmlFuncs = htmlTpl.FuncMap{
		"fname": templateGetFirstName,
	}

	textFuncs = textTpl.FuncMap{
		"fname": templateGetFirstName,
	}
}

func renderTextTemplate(filePath string, data templateData) (string, error) {
	tpl := textTpl.Must(textTpl.New(path.Base(filePath)).Funcs(textFuncs).ParseFiles(filePath))
	var outBuffer bytes.Buffer
	err := tpl.Execute(&outBuffer, data)
	if err != nil {
		return "", err
	}
	return outBuffer.String(), nil
}

func renderHTMLTemplate(basePath string, filePath string, data templateData) (string, error) {
	tpl := htmlTpl.Must(htmlTpl.New(path.Base(filePath)).Funcs(htmlFuncs).ParseFiles(filePath, basePath))
	var outBuffer bytes.Buffer
	err := tpl.Execute(&outBuffer, data)
	if err != nil {
		return "", err
	}
	return outBuffer.String(), nil
}

// Send sends the email with the given template name and data to the specified recipient. If recipient is blank, the user's email is used as the recipient.
func Send(recipient string, user *data.User, templateName string, data map[string]interface{}) error {
	emailConfig := config.GetCurrent().Email

	if !emailConfig.Enabled {
		return ErrDisabled
	}

	toAddress := recipient
	toFull := recipient
	if toAddress == "" {
		toAddress = user.Email
		if user.Name != "" {
			toFull = fmt.Sprintf("%s <%s>", user.Name, user.Email)
		} else {
			toFull = user.Email
		}
	}

	dot := templateData{user, toAddress, data}

	// use templates
	templateLocation := "templates/"
	basePath := filepath.Join(templateLocation, "base.html")
	baseFolder := filepath.Join(templateLocation, templateName)

	subject, err := renderTextTemplate(filepath.Join(baseFolder, "subject.txt"), dot)
	if err != nil {
		return err
	}

	textMessage, err := renderTextTemplate(filepath.Join(baseFolder, "template.txt"), dot)
	if err != nil {
		return err
	}

	htmlMessage, err := renderHTMLTemplate(basePath, filepath.Join(baseFolder, "template.html"), dot)
	if err != nil {
		return err
	}

	// encode as quotedprintable
	htmlEncodedBuf := bytes.NewBufferString("")
	htmlEncodedWriter := quotedprintable.NewWriter(htmlEncodedBuf)
	htmlEncodedWriter.Write([]byte(htmlMessage))
	htmlEncodedWriter.Close()

	// generate message id
	messageIDFirst := strconv.FormatInt(time.Now().Unix(), 10) + "." + strconv.Itoa(rand.Intn(999999))
	messageIDDomain := strings.Split(emailConfig.SMTPUsername, "@")[1]

	rawMessage := []byte(
		"From: " + emailConfig.From + "\r\n" +
			"To: " + toFull + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Sender: " + emailConfig.From + "\r\n" +
			"Message-ID: <" + messageIDFirst + "@" + messageIDDomain + ">\r\n" +
			"Date: " + time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700") + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: multipart/alternative; boundary=\"mimeboundary\"\r\n\r\n" +
			"--mimeboundary\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			textMessage + "\r\n" +
			"--mimeboundary\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
			"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
			htmlEncodedBuf.String() + "\r\n" +
			"--mimeboundary--")

	auth := smtp.PlainAuth("", emailConfig.SMTPUsername, emailConfig.SMTPPassword, emailConfig.SMTPHost)

	err = smtp.SendMail(emailConfig.SMTPHost+":"+strconv.Itoa(emailConfig.SMTPPort), auth, emailConfig.From, []string{toAddress}, rawMessage)
	if err != nil {
		return err
	}

	return nil
}
