package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialogManager"
)

func GetTextInputProcessorManager() dialogManager.TextInputProcessorManager {
	return dialogManager.TextInputProcessorManager {
		Processors : dialogManager.TextProcessorsMap {
			// "newTimezone" : processSetTimezone,
		},
	}
}

// func processSetTimezone(additionalId int64, data *processing.ProcessData) bool {
// 	_, err := time.LoadLocation(data.Message)

// 	if err == nil {
// 		staticFunctions.GetDb(data.Static).SetUserTimezone(data.UserId, data.Message)
// 		data.SendDialog(data.Static.MakeDialogFn("us", data.UserId, data.Trans, data.Static))
// 		return true
// 	} else {
// 		data.Static.SetUserStateTextProcessor(data.UserId, &processing.AwaitingTextProcessorData{
// 			ProcessorId: "newTimezone",
// 		})
// 		data.SendMessage(data.Trans("wrong_timezone") + "\n" + data.Trans("send_timezone"))
// 		return true
// 	}
// }
