package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-spy-game-bot/staticData"
	"github.com/nicksnyder/go-i18n/i18n"
	"log"
	"math/rand"
	"strings"
)

func SendThemeToPlayers(staticData *processing.StaticProccessStructs, userIds []int64, theme string) (success bool) {
	db := GetDb(staticData)

	if len(userIds) < 2 {
		return false
	}

	spyIdx := rand.Intn(len(userIds))

	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)

		var themeMessage string
		if i == spyIdx {
			themeMessage = "<tg-spoiler>" + trans("theme_spy") + "</tg-spoiler>"
		} else {
			themeMessage = theme
		}

		chatId, isFound := db.GetTelegramUserChatId(userId)
		if isFound {
			staticData.Chat.SendMessage(chatId, wrapIntoTelegramSpoiler(themeMessage, trans), 0, true)
		} else {
			db.AddWebMessage(userId, themeMessage, 10)
		}
	}
	return true
}

func SendThemeToOthers(staticData *processing.StaticProccessStructs, sessionId int64, excludeUserId int64, theme string) (success bool) {
	playersInSession := GetDb(staticData).GetUsersInSession(sessionId)
	var playersExceptCurrent []int64
	for _, userId := range playersInSession {
		if userId != excludeUserId {
			playersExceptCurrent = append(playersExceptCurrent, userId)
		}
	}

	return SendThemeToPlayers(staticData, playersExceptCurrent, theme)
}

func SendSpyfallLocationToAll(staticData *processing.StaticProccessStructs, sessionId int64) (success bool) {
	db := GetDb(staticData)

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		log.Print("Config type is incorrect")
		return false
	}

	locationsCount := len(config.SpyfallLocations)

	if locationsCount == 0 {
		log.Print("No locations found")
		return false
	}

	locationIdx := rand.Intn(locationsCount)
	locationInfoCopy := config.SpyfallLocations[locationIdx]
	rand.Shuffle(len(locationInfoCopy.Roles), func(i, j int) {
		locationInfoCopy.Roles[i], locationInfoCopy.Roles[j] = locationInfoCopy.Roles[j], locationInfoCopy.Roles[i]
	})

	userIds := db.GetUsersInSession(sessionId)

	if len(userIds) < 2 {
		return false
	}

	spyIdx := rand.Intn(len(userIds))

	roleIdx := 0
	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)

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

		chatId, isFound := db.GetTelegramUserChatId(userId)
		if isFound {
			staticData.Chat.SendMessage(chatId, wrapIntoTelegramSpoiler(theme, trans), 0, true)
		} else {
			db.AddWebMessage(userId, theme, 10)
		}
	}
	return true
}

func wrapIntoTelegramSpoiler(text string, trans i18n.TranslateFunc) string {
	// since Telegram spoiler tag shows length, add spaces at the end of each line to obfuscate it
	var finalText string
	const maxLineLength = 40
	for {
		separatorPos := strings.Index(text, "\n")
		if separatorPos == -1 {
			finalText += text + "\n"
			break
		}

		// skip long lines
		if separatorPos > maxLineLength {
			finalText += text[separatorPos+1:]
			text = text[separatorPos+1:]
			continue
		}

		finalText += text[:separatorPos] + strings.Repeat(" ", maxLineLength-len(text[:separatorPos])) + "\n"
		text = text[separatorPos+1:]
	}
	// add an extra static line, since spaces from the last line are cut off
	return "<tg-spoiler>" + finalText + "<i>" + trans("spoiler_terminator") + "</i>" + "</tg-spoiler>"
}

func SendSpyfallLocationsList(data *processing.ProcessData) {
	config, configCastSuccess := data.Static.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		log.Print("Config type is incorrect")
		return
	}

	var themesList []string
	for _, location := range config.SpyfallLocations {
		themesList = append(themesList, data.Trans("spyfall_loc_"+location.LocationId))
	}

	data.SendMessage(strings.Join(themesList[:], "\n"), true)
}

func GiveRandomNumbersToPlayers(staticData *processing.StaticProccessStructs, sessionId int64) {
	db := GetDb(staticData)

	userIds := db.GetUsersInSession(sessionId)

	if len(userIds) < 2 {
		return
	}

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })
	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)
		theme := trans("player_number_msg", map[string]interface{}{
			"Number": i + 1,
		})

		chatId, isFound := db.GetTelegramUserChatId(userId)
		if isFound {
			staticData.Chat.SendMessage(chatId, theme, 0, true)
		} else {
			db.AddWebMessage(userId, theme, 10)
		}
	}
	return
}
