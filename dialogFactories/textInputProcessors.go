package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialogManager"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
)

func GetTextInputProcessorManager() dialogManager.TextInputProcessorManager {
	return dialogManager.TextInputProcessorManager{
		Processors: dialogManager.TextProcessorsMap{
			//"connectSession" : processConnectSession,
		},
	}
}

func processConnectSession(additionalId int64, data *processing.ProcessData) bool {
	isSuccessful := staticFunctions.ConnectToSession(data, data.Message)
	if isSuccessful {
		return true
	}

	data.Static.SetUserStateTextProcessor(data.UserId, &processing.AwaitingTextProcessorData{
		ProcessorId: "connectSession",
	})
	data.SendMessage(data.Trans("session_not_found_try_again"), true)
	return true
}
