package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-spy-game-bot/staticData"
	"log"
	"math/rand"
	"strings"
)

func SendThemeToPlayers(data *processing.ProcessData, userIds []int64, theme string) (success bool) {
	db := GetDb(data.Static)

	if len(userIds) < 2 {
		data.SendMessage(data.Trans("few_players"), true)
		return false
	}

	spyIdx := rand.Intn(len(userIds))

	for i, userId := range userIds {
		trans := FindTransFunction(userId, data.Static)

		sentTheme := theme
		if i == spyIdx {
			sentTheme = trans("theme_spy")
		}

		db.SetThemeRevealed(userId, false)
		db.SetUserTheme(userId, sentTheme)
		chatId := db.GetChatId(userId)
		data.Static.Chat.SendDialog(chatId, data.Static.MakeDialogFn("th", userId, trans, data.Static, sentTheme), 0)
	}
	return true
}

func SendThemeToOthers(data *processing.ProcessData, sessionId int64, theme string) {
	playersInSession := GetDb(data.Static).GetUsersInSession(sessionId)
	playersExceptCurrent := []int64{}
	for _, userId := range playersInSession {
		if userId != data.UserId {
			playersExceptCurrent = append(playersExceptCurrent, userId)
		}
	}

	success := SendThemeToPlayers(data, playersExceptCurrent, theme)

	if success {
		data.SendMessage(data.Trans("theme_sent"), true)
	}
}

func SendSpyfallLocationToAll(data *processing.ProcessData, sessionId int64) (success bool) {
	db := GetDb(data.Static)

	config, configCastSuccess := data.Static.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		log.Print("Config type is incorrect")
		return false
	}

	locationIdx := rand.Intn(len(config.SpyfallLocations))
	locationInfoCopy := config.SpyfallLocations[locationIdx]
	rand.Shuffle(len(locationInfoCopy.Roles), func(i, j int) {
		locationInfoCopy.Roles[i], locationInfoCopy.Roles[j] = locationInfoCopy.Roles[j], locationInfoCopy.Roles[i]
	})

	userIds := db.GetUsersInSession(sessionId)

	if len(userIds) < 2 {
		data.SendMessage(data.Trans("few_players"), true)
		return false
	}

	spyIdx := rand.Intn(len(userIds))

	roleIdx := 0
	for i, userId := range userIds {
		trans := FindTransFunction(userId, data.Static)

		var theme string
		if i == spyIdx {
			theme = trans("spyfall_theme_spy")
		} else if roleIdx < len(locationInfoCopy.Roles)-1 {
			theme = trans("spyfall_theme", map[string]interface{}{
				"Location": trans("spyfall_loc_" + locationInfoCopy.LocationId),
				"Role":     trans("spyfall_role_" + locationInfoCopy.LocationId + "_" + locationInfoCopy.Roles[roleIdx]),
			})
			roleIdx += 1
		}

		db.SetThemeRevealed(userId, false)
		db.SetUserTheme(userId, theme)
		chatId := db.GetChatId(userId)
		data.Static.Chat.SendDialog(chatId, data.Static.MakeDialogFn("th", userId, trans, data.Static, theme), 0)
	}
	return true
}

func SendSpyfallLocationsList(data *processing.ProcessData) {
	config, configCastSuccess := data.Static.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		log.Print("Config type is incorrect")
		return
	}

	themesList := []string{}
	for _, location := range config.SpyfallLocations {
		themesList = append(themesList, data.Trans("spyfall_loc_"+location.LocationId))
	}

	data.SendMessage(strings.Join(themesList[:], "\n"), true)
}

func GiveRandomNumbersToPlayers(data *processing.ProcessData, sessionId int64) {
	db := GetDb(data.Static)

	userIds := db.GetUsersInSession(sessionId)

	if len(userIds) < 2 {
		data.SendMessage(data.Trans("few_players"), true)
		return
	}

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })
	for i, userId := range userIds {
		trans := FindTransFunction(userId, data.Static)
		theme := trans("player_number_msg", map[string]interface{}{
			"Number": i + 1,
		})

		chatId := db.GetChatId(userId)
		data.Static.Chat.SendMessage(chatId, theme, 0, true)
	}
	return
}
