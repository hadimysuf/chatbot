package route

import (
	"log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/controller"
	"github.com/gocroot/helper"
)

func URL(w http.ResponseWriter, r *http.Request) {
	if config.ErrorMongoconn != nil {
		log.Println(config.ErrorMongoconn.Error())
	}

	var method, path string = r.Method, r.URL.Path
	switch {
	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	case method == "GET" && path == "/refresh/token/lmsdesa":
		controller.GetNewTokenLMSDesa(w, r)
	case method == "GET" && path == "/refresh/token":
		controller.GetNewToken(w, r)
	case method == "POST" && helper.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)
	case method == "POST" && helper.URLParam(path, "/webhook/telebot/:nomorwa"):
		controller.TelebotWebhook(w, r)
	case method == "GET" && path == "/strava/activities":
		controller.GetStravaActivities(w, r)
	case method == "GET" && path == "/data/stravaactivities":
		controller.GetStravaActivitiesWithGrupIDFromPomokit(w, r)
	case method == "GET" && path == "/data/pomokit":
		controller.GetPomokitData(w, r)
	case method == "GET" && helper.URLParam(path, "/data/pomokit/:nomorwa"):
		controller.GetPomokitDataByPhonenumber(w, r)
	default:
		controller.NotFound(w, r)
	}
}
