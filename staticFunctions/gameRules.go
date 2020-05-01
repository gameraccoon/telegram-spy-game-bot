package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
)

func SendThemeToPlayers(staticData *processing.StaticProccessStructs, userIds []int64, theme string) {
	db := GetDb(staticData)
	for _, userId := range userIds {
		db.SetThemeRevealed(userId, false)
		db.SetUserTheme(userId, theme)
		trans := FindTransFunction(userId, staticData)
		chatId := db.GetChatId(userId)
		staticData.Chat.SendDialog(chatId, staticData.MakeDialogFn("th", userId, trans, staticData, theme), 0)
	}
}

func SendThemeToOthers(data *processing.ProcessData, sessionId int64, theme string) {
	playersInSession := GetDb(data.Static).GetUsersInSession(sessionId)
	playersExceptCurrent := []int64{}
	for _, userId := range playersInSession {
		playersExceptCurrent = append(playersExceptCurrent, userId)
	}
	SendThemeToPlayers(data.Static, playersExceptCurrent, theme)
	data.SendMessage(data.Trans("theme_sent"))
}

func SendThemeToAll(data *processing.ProcessData, sessionId int64, theme string) {
	playersInSession := GetDb(data.Static).GetUsersInSession(sessionId)
	SendThemeToPlayers(data.Static, playersInSession, theme)
}