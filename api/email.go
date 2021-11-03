package api

import (
	"archive/zip"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"oauth2-microsoft/database"
	"oauth2-microsoft/model"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SendEmail will send an email using the api
func SendEmail(user model.User, email model.SendEmailStruct) (string, error) {

	data, _ := json.Marshal(email)
	return CallAPIMethod("POST", "/me/sendMail", user.AccessToken, "", []byte(data), "application/json")
	//log.Printf("E-mail to : %s responded with status code: %d", email.Message.ToRecipients[0].EmailAddress.Address, respcode)
}

func GetEmailById(user model.User, emailID string) model.SingleMail {

	additionalParameters := url.Values{}
	additionalParameters.Add("select", "receivedDateTime,hasAttachments,importance,subject,sender,bodyPreview,body")
	endpoint := fmt.Sprintf("/me/messages/%s?", emailID)
	messageResponse, _ := CallAPIMethod("GET", endpoint, user.AccessToken, additionalParameters.Encode(), nil, "")
	mail := model.SingleMail{}
	json.Unmarshal([]byte(messageResponse), &mail)
	return mail

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func DownloadAttachment(user model.User, mailId string) []model.AttachmentDb{
	var arr []model.AttachmentDb
	folderDir := fmt.Sprintf("./attachment/%s", mailId)
	if _, err := os.Stat(folderDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(folderDir, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
	}
	endPoint := fmt.Sprintf("/me/messages/%s/attachments", mailId)
	messagesResponse, _ := CallAPIMethod("GET", endPoint, user.AccessToken, "", nil, "")
	//fmt.Println(messagesResponse)
	var msgAttachments model.MsgAttachments
	json.Unmarshal([]byte(messagesResponse), &msgAttachments)
	for _, v := range msgAttachments.Value {
		downFile := fmt.Sprintf("%s/%s", folderDir, filepath.Base(v.Name))
		arr = append(arr, model.AttachmentDb{
			AttachmentId: v.Id,
			MailId:       mailId,
			Filename:     v.Name,
			Filepath:     downFile,
		})
		if ok, err := PathExists(downFile); err == nil && ok {
			continue
		}
		decodeBytes, err := b64.StdEncoding.DecodeString(v.ContentBytes)
		if err!=nil {
			log.Println(err)
			continue
		}
		ioutil.WriteFile(downFile, decodeBytes, 0644)

	}
	return arr
}
func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()
	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// 打包成zip文件
func Zip(src_dir string, zip_file_name string) {
	// 预防：旧文件无法覆盖
	os.RemoveAll(zip_file_name)
	// 创建：zip文件
	zipfile, _ := os.Create(zip_file_name)
	defer zipfile.Close()
	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	// 遍历路径信息
	filepath.Walk(src_dir, func(path string, info os.FileInfo, _ error) error {
		// 如果是源路径，提前进行下一个遍历
		if path == src_dir {
			return nil
		}
		// 获取：文件头信息
		header, _ := zip.FileInfoHeader(info)
		header.Name = strings.TrimPrefix(path, src_dir+`/`)

		// 判断：文件是不是文件夹
		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, _ := archive.CreateHeader(header)
		if !info.IsDir() {
			file, _ := os.Open(path)
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})
}

func GenerateMailHtml(user model.User, mails []model.Mail) (string, error) {
	folderDir := fmt.Sprintf("./exports/%s/%s", user.UserPrincipalName, fmt.Sprintf("%s", time.Now().Format("2006_01_02_15_04_05")))
	folderHtmlDir := fmt.Sprintf("%s/html", folderDir)

	if _, err := os.Stat(folderHtmlDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(folderHtmlDir, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
	}

	for _, v := range mails {
		filename := fmt.Sprintf("%s/%s.html", folderHtmlDir,  v.Id)
		ioutil.WriteFile(filename, []byte(v.BodyContent), 0644)
		for _, item := range v.Attachments {
			mailAttachDir := fmt.Sprintf("%s/attachment/%s", folderDir, item.MailId)
			os.MkdirAll(mailAttachDir, os.ModePerm)
			dest := fmt.Sprintf("%s/%s", mailAttachDir, item.Filename)
			copyFile(item.Filepath, dest)

		}
	}
	page := model.Page{
		Title:             "",
		Email:             "",
		User:              model.User{},
		UserList:          nil,
		EmailList:         mails,
		FileList:          nil,
		SearchFiles:       model.Files{},
		Mail:              model.SingleMail{},
		Message:           "",
		Success:           false,
		URL:               "",
	}
	tpl, err := template.ParseFiles("templates/exportcatalogue.html")
	if err != nil {
		log.Fatal(err)
	}
	indexFileName := fmt.Sprintf("%s/index.html", folderDir)
	file, err := os.Create(indexFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	tpl.Execute(file,  page)
	// 压缩
	zipFile := fmt.Sprintf("%s.zip", folderDir)
	Zip(folderDir, zipFile)
	//os.RemoveAll(folderDir)
	return zipFile, nil
}


func GetAllEmails(user model.User, insertInDB bool) []model.Mail {

	//keyWords := strings.Split(searchKeywords, ",")
	additionalParameters := url.Values{}
	additionalParameters.Add("select", "receivedDateTime,hasAttachments,importance,subject,sender,bodyPreview,body,toRecipients")
	messagesResponse, _ := CallAPIMethod("GET", "/me/messages?", user.AccessToken, additionalParameters.Encode(), nil, "")
	messages := model.Messages{}
	json.Unmarshal([]byte(messagesResponse), &messages)

	countEffectiveMsg := 0
	countBatchNotInDb := 0
	// Loads the first batch of emails.
	for _, message := range messages.Value {
		if len(message.ToRecipients) > 0 {
			countEffectiveMsg++
			if database.GetEmail(message.ID) == nil {
				var watts []model.AttachmentDb
				if message.HasAttachments {
					watts = DownloadAttachment(user, message.ID)
				}
				oneMail := model.Mail{
					Id:              message.ID,
					User:            user.Mail,
					Subject:         message.Subject,
					SenderEmail:     message.Sender.EmailAddress.Address,
					SenderName:      message.Sender.EmailAddress.Name,
					Attachments:     watts,
					BodyPreview:     message.BodyPreview,
					BodyType:        message.Body.ContentType,
					BodyContent:     message.Body.Content,
					ToRecipient:     message.ToRecipients[0].EmailAddress.Address,
					ToRecipientName: message.ToRecipients[0].EmailAddress.Name,
					Date:            message.ReceivedDateTime}
				database.InsertEmail(oneMail)
				countBatchNotInDb++
			}
		}
	}
	// 如果这个批次的数据都不存在于数据库中，才进行迭代查询，否则直接数据库查询
	if countBatchNotInDb == countEffectiveMsg {
		for messages.OdataNextLink != "" {
			countBatchNotInDb = 0
			countEffectiveMsg = 0
			endpoint := strings.Replace(messages.OdataNextLink, model.ApiEndpointRoot, "", -1)
			messagesResponse, _ = CallAPIMethod("GET", endpoint, user.AccessToken, "", nil, "")
			messages = model.Messages{}
			json.Unmarshal([]byte(messagesResponse), &messages)
			// Load next batch of emails
			for _, message := range messages.Value {
				if len(message.ToRecipients) > 0 {
					countEffectiveMsg++
					if database.GetEmail(message.ID) == nil {
						var watts []model.AttachmentDb
						if message.HasAttachments {
							watts = DownloadAttachment(user, message.ID)
						}
						oneMail := model.Mail{
							Id:              message.ID,
							User:            user.Mail,
							Subject:         message.Subject,
							SenderEmail:     message.Sender.EmailAddress.Address,
							SenderName:      message.Sender.EmailAddress.Name,
							Attachments:     watts,
							BodyPreview:     message.BodyPreview,
							BodyType:        message.Body.ContentType,
							BodyContent:     message.Body.Content,
							ToRecipient:     message.ToRecipients[0].EmailAddress.Address,
							ToRecipientName: message.ToRecipients[0].EmailAddress.Name,
							Date:            message.ReceivedDateTime}
						database.InsertEmail(oneMail)
						countBatchNotInDb++
					}
				}
			}
			if countBatchNotInDb == countEffectiveMsg {
				continue
			} else {
				return database.GetEmailsByUser(user.UserPrincipalName)
			}
		}
	}
	return database.GetEmailsByUser(user.UserPrincipalName)
}
//
//func GetKeywordEmails(user model.User, searchKeyword string, insertInDB bool) []model.Mail {
//
//	dbMails := []model.Mail{}
//
//	//keyWords := strings.Split(searchKeywords, ",")
//
//	additionalParameters := url.Values{}
//	additionalParameters.Add("select", "receivedDateTime,hasAttachments,importance,subject,sender,bodyPreview,body,toRecipients")
//	additionalParameters.Add("$search", searchKeyword)
//
//	messagesResponse, _ := CallAPIMethod("GET", "/me/messages?", user.AccessToken, additionalParameters.Encode(), nil, "")
//	messages := model.Messages{}
//
//	json.Unmarshal([]byte(messagesResponse), &messages)
//
//	// Loads the first batch of emails.
//	for _, message := range messages.Value {
//		dbMails = append(dbMails, model.Mail{
//			Id:              message.ID,
//			User:            user.Mail,
//			Subject:         message.Subject,
//			SenderEmail:     message.Sender.EmailAddress.Address,
//			SenderName:      message.Sender.EmailAddress.Name,
//			HasAttachments:  message.HasAttachments,
//			BodyPreview:     message.BodyPreview,
//			BodyType:        message.Body.ContentType,
//			BodyContent:     message.Body.Content,
//			ToRecipient:     message.ToRecipients[0].EmailAddress.Address,
//			ToRecipientName: message.ToRecipients[0].EmailAddress.Name, Date: message.ReceivedDateTime})
//	}
//
//	for messages.OdataNextLink != "" {
//		endpoint := strings.Replace(messages.OdataNextLink, model.ApiEndpointRoot, "", -1)
//		//fmt.Println(endpoint)
//		messagesResponse, _ = CallAPIMethod("GET", endpoint, user.AccessToken, "", nil, "")
//
//		messages = model.Messages{}
//		json.Unmarshal([]byte(messagesResponse), &messages)
//		// Load next batch of emails
//		for _, message := range messages.Value {
//			dbMails = append(dbMails, model.Mail{
//				Id:              message.ID,
//				User:            user.Mail,
//				Subject:         message.Subject,
//				SenderEmail:     message.Sender.EmailAddress.Address,
//				SenderName:      message.Sender.EmailAddress.Name,
//				HasAttachments:  message.HasAttachments,
//				BodyPreview:     message.BodyPreview,
//				BodyType:        message.Body.ContentType,
//				BodyContent:     message.Body.Content,
//				ToRecipient:     message.ToRecipients[0].EmailAddress.Address,
//				ToRecipientName: message.ToRecipients[0].EmailAddress.Name, Date: message.ReceivedDateTime})
//		}
//	}
//
//	if insertInDB {
//		log.Printf("Inserting %d keyworded emails from %s", len(dbMails), user.Mail)
//		for _, mail := range dbMails {
//			database.InsertEmail(mail)
//		}
//	}
//	return dbMails
//
//}
