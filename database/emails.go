package database

import (
	_ "database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"oauth2-microsoft/model"
)

func GetEmailsByUser(email string) []model.Mail{

	var mails []model.Mail
	rows, err := db.Query(model.GetUserMailsQuery,email)
	mail := model.Mail{}

	if err != nil{
		log.Println("Error : " + err.Error())
	}
	for rows.Next() {
		var temp string
		err := rows.Scan(&mail.Id,
			&mail.User,
			&mail.Subject,
			&mail.SenderEmail,
			&mail.SenderName,
			&temp,
			&mail.BodyPreview,
			&mail.BodyType,
			&mail.BodyContent,
			&mail.ToRecipient,
			&mail.ToRecipientName,
			&mail.Date)
		json.Unmarshal([]byte(temp), &mail.Attachments)
		if err != nil {
			log.Fatal(err)
		}
		mails = append(mails,mail)
	}

	return mails
}

//func GetAllEmails() []model.Mail{
//
//	var mails []model.Mail
//
//
//
//	rows, err := db.Query(model.GetMailsQuery)
//	mail := model.Mail{}
//
//	if err != nil{
//		log.Println("Error : " + err.Error())
//	}
//	for rows.Next() {
//		err := rows.Scan(&mail.Id,&mail.User,&mail.Subject,&mail.SenderEmail,&mail.SenderName,&mail.HasAttachments,&mail.BodyPreview,&mail.BodyType,&mail.BodyContent)
//		if err != nil {
//			log.Fatal(err)
//		}
//		mails = append(mails,mail)
//	}
//	return mails
//}

func InsertEmail(mail model.Mail){

	tx, _ := db.Begin()
	stmt, err_stmt := tx.Prepare(model.InsertMailQuery)

	if err_stmt != nil {
		log.Fatal(err_stmt)
	}
	marshal, _ := json.Marshal(mail.Attachments)
	_, err := stmt.Exec(mail.Id,
		mail.User,
		mail.Subject,
		mail.SenderEmail,
		mail.SenderName,
		string(marshal),
		mail.BodyPreview,
		mail.BodyType,
		mail.BodyContent,
		mail.ToRecipient,
		mail.ToRecipientName,
		mail.Date)
	tx.Commit()
	if err != nil{
		log.Printf("ERROR: %s",err)
	}

}
//
//func SearchUserEmails(email string,searchKey string) []model.Mail {
//	var mails []model.Mail
//
//	searchKey = "%" + searchKey + "%"
//
//	rows, err := db.Query(model.SearchUserMailsQuery,email,searchKey)
//	mail := model.Mail{}
//
//	if err != nil{
//		log.Println("Error : " + err.Error())
//	}
//	for rows.Next() {
//		err := rows.Scan(&mail.Id,&mail.User,&mail.Subject,&mail.SenderEmail,&mail.SenderName,&mail.HasAttachments,&mail.BodyPreview,&mail.BodyType,&mail.BodyContent)
//		if err != nil {
//			log.Fatal(err)
//		}
//		mails = append(mails,mail)
//	}
//
//	return mails
//}
//
//
//func SearchEmails(searchKey string) []model.Mail {
//	var mails []model.Mail
//
//	searchKey = "%" + searchKey + "%"
//
//	rows, err := db.Query(model.SearchEmailQuery,searchKey)
//	mail := model.Mail{}
//
//	if err != nil{
//		log.Println("Error : " + err.Error())
//	}
//	for rows.Next() {
//		err := rows.Scan(
//			&mail.Id,
//			&mail.User,
//			&mail.Subject,
//			&mail.SenderEmail,
//			&mail.SenderName,
//			&mail.HasAttachments,
//			&mail.BodyPreview,
//			&mail.BodyType,
//			&mail.BodyContent)
//		if err != nil {
//			log.Fatal(err)
//		}
//		mails = append(mails,mail)
//	}
//
//	return mails
//}


func GetEmail(id string) *model.Mail {

	row := db.QueryRow(model.GetEmailQuery,id)
	mail := model.Mail{}
	var temp string
	err := row.Scan(&mail.Id,
		&mail.User,
		&mail.Subject,
		&mail.SenderEmail,
		&mail.SenderName,
		&temp,
		&mail.BodyPreview,
		&mail.BodyType,
		&mail.BodyContent,
		&mail.ToRecipient,
		&mail.ToRecipientName,
		&mail.Date)
	json.Unmarshal([]byte(temp), &mail.Attachments)
	if err != nil {
		return nil
	}

	return &mail
}
