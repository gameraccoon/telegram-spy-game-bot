package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-spy-game-bot/database"
	static "github.com/gameraccoon/telegram-spy-game-bot/staticData"
	"github.com/nicksnyder/go-i18n/i18n"
	"log"
	"strings"
	"time"
)

func GetDb(staticData *processing.StaticProccessStructs) *database.SpyBotDb {
	if staticData == nil {
		log.Fatal("staticData is nil")
		return nil
	}

	db, ok := staticData.Db.(*database.SpyBotDb)
	if ok && db != nil {
		return db
	} else {
		log.Fatal("database is not set properly")
		return nil
	}
}

func getClosestLang(config *static.StaticConfiguration, lang string) string {
	for _, langCode := range config.AvailableLanguages {
		if strings.HasPrefix(langCode.Key, lang) {
			return langCode.Key
		}
	}
	return lang
}

func FindTransFunction(userId int64, staticData *processing.StaticProccessStructs) i18n.TranslateFunc {
	// ToDo: cache user's lang
	lang := GetDb(staticData).GetUserLanguage(userId)

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	// replace empty language to default one (some clients don't send user's language)
	if len(lang) <= 0 {
		log.Printf("User %d has empty language. Setting to default.", userId)
		lang = config.DefaultLanguage
	}

	if foundTrans, ok := staticData.Trans[lang]; ok {
		GetDb(staticData).SetUserLanguage(userId, lang)
		return foundTrans
	}

	if foundTrans, ok := staticData.Trans[getClosestLang(&config, lang)]; ok {
		GetDb(staticData).SetUserLanguage(userId, lang)
		return foundTrans
	}

	// unknown language, use default instead
	if foundTrans, ok := staticData.Trans[config.DefaultLanguage]; ok {
		log.Printf("User %d has unknown language (%s). Setting to default.", userId, lang)
		lang = config.DefaultLanguage
		GetDb(staticData).SetUserLanguage(userId, lang)
		return foundTrans
	}

	// something gone wrong
	log.Printf("Translator didn't found: %s", lang)
	// fall to the first available translator
	for lang, trans := range staticData.Trans {
		log.Printf("Using first available translator: %s", lang)
		return trans
	}

	// something gone completely wrong
	log.Fatal("There are no available translators")
	// we will probably crash but there is nothing else we can do
	translator, _ := i18n.Tfunc(config.DefaultLanguage)
	return translator
}

func FormatTimestamp(timestamp time.Time, timezone string) string {
	// the list of timezones https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	loc, err := time.LoadLocation(timezone)
	if err == nil {
		return timestamp.In(loc).Format("15:04:05 _2.01.2006")
	} else {
		return timestamp.Format("15:04:05 _2.01.2006")
	}
}

func SendSessionDialog(data *processing.ProcessData) {
	messageId := data.SendDialog(data.Static.MakeDialogFn("se", data.UserId, data.Trans, data.Static, nil))
	GetDb(data.Static).SetSessionMessageId(data.UserId, messageId)
}

func UpdateSessionDialogs(sessionId int64, staticData *processing.StaticProccessStructs) {
	users := GetDb(staticData).GetUsersInSession(sessionId)
	db := GetDb(staticData)

	for _, userId := range users {
		trans := FindTransFunction(userId, staticData)
		dialog := staticData.MakeDialogFn("se", userId, trans, staticData, nil)
		chatId := db.GetChatId(userId)
		messageId := db.GetSessionMessageId(userId)
		staticData.Chat.SendDialog(chatId, dialog, messageId)
	}
}

func ConnectToSession(data *processing.ProcessData, token string) (successfull bool) {
	db := GetDb(data.Static)
	sessionId, isFound := db.GetSessionIdFromToken(token)
	if !isFound {
		return false
	}

	successfullyConnected, previousSessionId, wasInSession := db.ConnectToSession(data.UserId, sessionId)
	if !successfullyConnected {
		return false
	}

	SendSessionDialog(data)

	UpdateSessionDialogs(sessionId, data.Static)

	if wasInSession {
		UpdateSessionDialogs(previousSessionId, data.Static)
	}

	return true
}
