package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"math/rand"
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

func SendThemeToAll(data *processing.ProcessData, sessionId int64, theme string) {
	playersInSession := GetDb(data.Static).GetUsersInSession(sessionId)
	SendThemeToPlayers(data.Static, playersInSession, theme)
}
