package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-spy-game-bot/staticData"
	"log"
	"math/rand"
	"strings"
)

func SendThemeToPlayers(staticData *processing.StaticProccessStructs, userIds []int64, theme string) {
	db := GetDb(staticData)

	spyIdx := rand.Intn(len(userIds))

	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)

		sentTheme := theme
		if i == spyIdx {
			sentTheme = trans("theme_spy")
		}

		db.SetThemeRevealed(userId, false)
		db.SetUserTheme(userId, sentTheme)
		chatId := db.GetChatId(userId)
		staticData.Chat.SendDialog(chatId, staticData.MakeDialogFn("th", userId, trans, staticData, sentTheme), 0)
	}
}

func SendThemeToOthers(data *processing.ProcessData, sessionId int64, theme string) {
	playersInSession := GetDb(data.Static).GetUsersInSession(sessionId)
	playersExceptCurrent := []int64{}
	for _, userId := range playersInSession {
		if userId != data.UserId {
			playersExceptCurrent = append(playersExceptCurrent, userId)
		}
	}
	SendThemeToPlayers(data.Static, playersExceptCurrent, theme)
	data.SendMessage(data.Trans("theme_sent"))
}

func SendSpyfallLocationToAll(staticData *processing.StaticProccessStructs, sessionId int64) {
	db := GetDb(staticData)

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		log.Print("Config type is incorrect")
		return
	}

	locationIdx := rand.Intn(len(config.SpyfallLocations))
	locationInfoCopy := config.SpyfallLocations[locationIdx]
	rand.Shuffle(len(locationInfoCopy.Roles), func(i, j int) { locationInfoCopy.Roles[i], locationInfoCopy.Roles[j] = locationInfoCopy.Roles[j], locationInfoCopy.Roles[i] })

	userIds := db.GetUsersInSession(sessionId)
	spyIdx := rand.Intn(len(userIds))

	roleIdx := 0
	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)

		var theme string
		if i == spyIdx {
			theme = trans("spyfall_theme_spy")
		} else if roleIdx < len(locationInfoCopy.Roles) - 1 {
			theme = trans("spyfall_theme", map[string]interface{} {
				"Location": trans("spyfall_loc_" + locationInfoCopy.LocationId),
				"Role": trans("spyfall_role_" + locationInfoCopy.LocationId + "_" + locationInfoCopy.Roles[roleIdx]),
			})
			roleIdx += 1
		}

		db.SetThemeRevealed(userId, false)
		db.SetUserTheme(userId, theme)
		chatId := db.GetChatId(userId)
		staticData.Chat.SendDialog(chatId, staticData.MakeDialogFn("th", userId, trans, staticData, theme), 0)
	}
}

func SendSpyfallLocationsList(data *processing.ProcessData) {
	config, configCastSuccess := data.Static.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		log.Print("Config type is incorrect")
		return
	}

	themesList := []string{}
	for _, location := range config.SpyfallLocations {
		themesList = append(themesList, data.Trans("spyfall_loc_" + location.LocationId))
	}

	data.SendMessage(strings.Join(themesList[:], "\n"))
}