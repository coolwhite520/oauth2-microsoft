package server

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/kataras/iris/v12/context"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"oauth2-microsoft/api"
	"oauth2-microsoft/database"
	"oauth2-microsoft/model"
	"os"
	"path/filepath"
)

// This will contain the template functions

func ExecuteTemplate(w http.ResponseWriter, page model.Page, templatePath string) {

	tpl, err := template.ParseFiles("templates/main.html", templatePath)
	if err != nil {
		log.Fatal(err)
	}
	tpl.ExecuteTemplate(w, "layout", page)
}

func ExecuteSingleTemplate(w http.ResponseWriter, page model.Page, templatePath string) {

	tpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Println(err)
	}

	err = tpl.Execute(w, page)
	if err != nil {
		log.Println(err)
	}

}

func GetUsers(ctx context.Context) {
	Page := model.Page{}
	Page.Title = "Users"
	Page.URL = api.GenerateURL()
	Page.UserList = database.GetUsers()
	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/users.html")
}

func GetUserFiles(ctx context.Context) {
	Page := model.Page{}
	Page.Title = "Files"
	email := ctx.Params().Get("email")
	folderDir := fmt.Sprintf("./downloads/%s", email)
	if _, err := os.Stat(folderDir); err != nil {
		if os.IsNotExist(err) {
			// Create the folder
			ctx.ResponseWriter().Write([]byte("No files exist for this user"))
			return
		} else {
			log.Println(err)
		}
	}
	var files []string
	err := filepath.Walk(folderDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		files = append(files, filepath.Base(path))
		return nil
	})
	if err != nil {
		ctx.ResponseWriter().Write([]byte("No files exist for this user"))

		log.Println(err)
		return
	}

	Page.Email = email
	Page.FileList = files
	ExecuteSingleTemplate(ctx.ResponseWriter(), Page, "templates/files.html")
}

func GetAbout(ctx context.Context) {
	Page := model.Page{}
	Page.Title = "About"
	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/about.html")
}

func GetEmail(ctx context.Context) {
	var Page model.Page
	emailID := ctx.Params().Get("email_id")
	userMail := ctx.Params().Get("id")
	user := database.GetUser(userMail)
	Page.Mail = api.GetEmailById(user, emailID) //database.GetEmail(email)
	ExecuteSingleTemplate(ctx.ResponseWriter(), Page, "templates/email.html")
}

//GetLiveMain will give the template
func GetLiveMain(ctx context.Context) {
	Page := model.Page{}
	Page.Title = "Live Interaction"
	id := ctx.Params().Get("id")
	Page.User = database.GetUser(id)
	api.RefreshAccessToken(&Page.User)
	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/live.html")
}

//GetLiveEmails will give the template
func GetLiveEmails(ctx context.Context) {

	Page := model.Page{}
	keyword := ctx.URLParam("keyword")
	id := ctx.Params().Get("id")
	// If keywords is empty it will search with the default keywords
	if keyword == "" {
		Page.User = database.GetUser(id)
		Page.Title = "Show All E-mails"
		Page.EmailList = api.GetAllEmails(Page.User, true)
	} else {
		Page.User = database.GetUser(id)
		Page.Title = fmt.Sprintf("Search result for : %s", keyword)
		Page.EmailList = api.GetKeywordEmails(Page.User, keyword, false)
	}
	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/emails.html")
}

//SendEmail will send an email to a specific address.
func SendEmail(ctx context.Context) {
	Page := model.Page{}
	id := ctx.Params().Get("id")

	Page.User = database.GetUser(id)

	if err := ctx.Request().ParseMultipartForm(5 * 1024); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)

		return
	}

	Page.Title = "Users success"

	email := model.SendEmailStruct{}

	email.Message.Subject = ctx.Request().FormValue("subject")
	email.Message.Body.ContentType = ctx.Request().FormValue("contentType")
	email.Message.Body.Content =ctx.Request().FormValue("message")

	// This code needs fixing .
	emailAddress := model.EmailAddress{Address: ctx.Request().FormValue("emailtarget")}
	target := model.ToRecipients{EmailAddress: emailAddress}
	recp := []model.ToRecipients{target}
	email.Message.ToRecipients = recp
	email.SaveToSentItems = "false"

	// Parse the File
	file, fileHandler, err := ctx.Request().FormFile("attachment")
	if err == nil {

		attachment := model.Attachment{}
		attachment.OdataType = "#microsoft.graph.fileAttachment"
		attachment.Name = fileHandler.Filename
		attachment.ContentType = fileHandler.Header["Content-Type"][0]

		// Load the attachment
		attachmentData, _ := ioutil.ReadAll(file)
		encAttachment := b64.StdEncoding.EncodeToString(attachmentData)

		attachment.ContentBytes = encAttachment
		email.Message.Attachments = []model.Attachment{attachment}
		defer file.Close()
	}

	resp, code := api.SendEmail(Page.User, email)
	if code == 202 {
		Page.Message = "E-mail was sent successfully"

		Page.Success = true
	} else {
		Page.Message = resp
	}
	fmt.Println(resp)

	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/message.html")
}

//GetLiveEmails will give the template
func GetLiveFiles(ctx context.Context) {
	Page := model.Page{}
	keyword := ctx.URLParam("keyword")
	id := ctx.Params().Get("id")
	Page.User = database.GetUser(id)

	if keyword == "" {
		Page.Title = "Last 10 modified files"
		Page.SearchFiles = api.GetKeywordFiles(Page.User, ".", "?$orderby=lastModifiedDateTime&$top=10")
	} else {
		Page.Title = fmt.Sprintf("Search result for : %s", keyword)
		Page.SearchFiles = api.GetKeywordFiles(Page.User, keyword, "?$orderby=lastModifiedDateTime&$top=100")
	}

	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/filesearch.html")
}

func DownloadFileHandler(ctx context.Context) {

	Page := model.Page{}
	id := ctx.Params().Get("id")
	fileid := ctx.Params().Get("fileid")
	Page.User = database.GetUser(id)
	api.LiveDownloadFile(Page.User, fileid)
	Page.Success = true
	Page.Message = "File Downloaded"
	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/message.html")
}

//UpdateFile will send an email to a specific address.
func ReplaceFile(ctx context.Context) {
	Page := model.Page{}
	id := ctx.Params().Get("id")
	fileid := ctx.Params().Get("fileid")
	Page.User = database.GetUser(id)
	//r.ParseForm()

	if err := ctx.Request().ParseMultipartForm(5 * 1024); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		return
	}

	// Parse the File
	file, fileHeader, _ := ctx.Request().FormFile("attachment")
	fileContent, _ := ioutil.ReadAll(file)
	fileContentType := fileHeader.Header["Content-Type"][0]
	resp, code := api.UpdateFile(Page.User, fileid, fileHeader.Filename, fileContent, fileContentType)

	if code == 200 {
		//	Page.Success = true
		Page.Message = "File Updated Successfully"
		Page.Success = true

	} else {
		Page.Message = resp
	}
	ExecuteTemplate(ctx.ResponseWriter(), Page, "templates/message.html")
}
