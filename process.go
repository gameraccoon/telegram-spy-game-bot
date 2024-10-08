package main

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialogManager"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"strings"
)

type ProcessorFunc func(*processing.ProcessData)

type ProcessorFuncMap map[string]ProcessorFunc

func startCommand(data *processing.ProcessData) {
	if len(data.Message) > 0 {
		isSuccessful := staticFunctions.ConnectToSession(data, data.Message)
		if isSuccessful {
			return
		}

		data.SendMessage(data.Trans("link_session_is_old"), true)
	} else {
		data.SendMessage(data.Trans("start_message"), true)
	}

	sessionCommand(data)
	data.Static.SetUserStateTextProcessor(data.UserId, nil)
}

func sessionCommand(data *processing.ProcessData) {
	_, isInSession := staticFunctions.GetDb(data.Static).GetUserSession(data.UserId)
	if isInSession {
		staticFunctions.SendSessionDialog(data)
	} else {
		data.SendDialog(data.Static.MakeDialogFn("ns", data.UserId, data.Trans, data.Static, nil))
	}
	data.Static.SetUserStateTextProcessor(data.UserId, nil)
}

func settingsCommand(data *processing.ProcessData) {
	data.SendDialog(data.Static.MakeDialogFn("us", data.UserId, data.Trans, data.Static, nil))
}

func sendSpyfallLocation(data *processing.ProcessData) {
	sessionId, isInSession := staticFunctions.GetDb(data.Static).GetUserSession(data.UserId)
	if isInSession {
		isSuccess := staticFunctions.SendSpyfallLocationToAll(data.Static, sessionId)
		if !isSuccess {
			trans := staticFunctions.FindTransFunction(data.UserId, data.Static)
			data.SendMessage(trans("few_players"), true)
		}
	} else {
		data.SendMessage(data.Trans("no_session_error"), true)
	}
}

func listOfSpyfallLocations(data *processing.ProcessData) {
	staticFunctions.SendSpyfallLocationsList(data)
}

func sendNumbersToPlayers(data *processing.ProcessData) {
	sessionId, isInSession := staticFunctions.GetDb(data.Static).GetUserSession(data.UserId)
	if isInSession {
		staticFunctions.GiveRandomNumbersToPlayers(data.Static, sessionId)
	} else {
		data.SendMessage(data.Trans("no_session_error"), true)
	}
}

func helpCommand(data *processing.ProcessData) {
	data.SendMessage(data.Trans("help_info"), true)
}

func cancelCommand(data *processing.ProcessData) {
	data.Static.SetUserStateTextProcessor(data.UserId, nil)
	data.SendMessage(data.Trans("command_canceled"), true)
}

func makeUserCommandProcessors() ProcessorFuncMap {
	return map[string]ProcessorFunc{
		"start":        startCommand,
		"session":      sessionCommand,
		"settings":     settingsCommand,
		"spyfall_send": sendSpyfallLocation,
		"spyfall_list": listOfSpyfallLocations,
		"help":         helpCommand,
		"cancel":       cancelCommand,
		"number":       sendNumbersToPlayers,
	}
}

func processCommandByProcessors(data *processing.ProcessData, processors *ProcessorFuncMap) bool {
	processor, ok := (*processors)[data.Command]
	if ok {
		processor(data)
	}

	return ok
}

func UpdateProcessData(data *processing.ProcessData) {
	userId := staticFunctions.GetDb(data.Static).GetOrCreateTelegramUserId(data.ChatId, data.UserSystemLang)
	data.UserId = userId
	data.Trans = staticFunctions.FindTransFunction(userId, data.Static)
}

func processCommand(data *processing.ProcessData, dialogManager *dialogManager.DialogManager, processors *ProcessorFuncMap) (succeeded bool) {
	UpdateProcessData(data)

	// drop any text processors for the case we will process a command
	data.Static.SetUserStateTextProcessor(data.UserId, nil)
	// process dialogs
	ids := strings.Split(data.Command, "_")
	if len(ids) >= 2 {
		dialogId := ids[0]
		variantId := ids[1]
		var additionalId string
		if len(ids) > 2 {
			additionalId = ids[2]
		}

		processed := dialogManager.ProcessVariant(dialogId, variantId, additionalId, data)
		if processed {
			return true
		}
	}

	// process static command
	processed := processCommandByProcessors(data, processors)
	if processed {
		return true
	}

	// if we here that means that no command was processed
	data.SendMessage(data.Trans("help_info"), true)
	return false
}

func processPlainMessage(data *processing.ProcessData, dialogManager *dialogManager.DialogManager) {
	UpdateProcessData(data)

	success := dialogManager.ProcessText(data)

	if !success {
		sessionId, isInSession := staticFunctions.GetDb(data.Static).GetUserSession(data.UserId)
		if isInSession {
			isSuccess := staticFunctions.SendThemeToOthers(data.Static, sessionId, data.UserId, data.Message)
			trans := staticFunctions.FindTransFunction(data.UserId, data.Static)
			if isSuccess {
				data.SendMessage(trans("theme_sent"), true)
			} else {
				data.SendMessage(trans("few_players"), true)
			}
		} else {
			data.SendMessage(data.Trans("help_info"), true)
		}
	}
}
