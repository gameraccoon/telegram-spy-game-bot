package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialogManager"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"strconv"
)

func GetTextInputProcessorManager() dialogManager.TextInputProcessorManager {
	return dialogManager.TextInputProcessorManager {
		Processors : dialogManager.TextProcessorsMap {
			"connectSession" : processConnectSession,
		},
	}
}

func processConnectSession(additionalId int64, data *processing.ProcessData) bool {
	sessionId, err := strconv.ParseInt(data.Message, 10, 64)
	if err == nil {
		isSuccessfull := staticFunctions.ConnectToSession(data, sessionId)
		if isSuccessfull {
			return true
		}
	}

	data.Static.SetUserStateTextProcessor(data.UserId, &processing.AwaitingTextProcessorData{
		ProcessorId: "connectSession",
	})
	data.SendMessage(data.Trans("session_not_found_try_again"))
	return true
}
