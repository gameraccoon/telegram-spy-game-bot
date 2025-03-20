package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"github.com/nicksnyder/go-i18n/i18n"
	"log"
	"strconv"
)

type sessionVariantPrototype struct {
	id         string
	textId     string
	process    func(int64, *processing.ProcessData) bool
	rowId      int
	isActiveFn func() bool
}

type sessionDialogFactory struct {
	variants []sessionVariantPrototype
}

func MakeSessionDialogFactory() dialogFactory.DialogFactory {
	return &(sessionDialogFactory{
		variants: []sessionVariantPrototype{
			sessionVariantPrototype{
				id:      "share",
				textId:  "share_link",
				process: openInviteDialog,
				rowId:   1,
			},
			sessionVariantPrototype{
				id:      "discsess",
				textId:  "disconnect_session",
				process: disconnectSession,
				rowId:   1,
			},
			sessionVariantPrototype{
				id:      "spyfallloc",
				textId:  "send_spyfall_location",
				process: sendSpyfallLocation,
				rowId:   2,
			},
		},
	})
}

func openInviteDialog(sessionId int64, data *processing.ProcessData) bool {
	data.SendDialog(data.Static.MakeDialogFn("in", data.UserId, data.Trans, data.Static, nil))
	return true
}

func disconnectSession(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	currentSessionId, isInSession := db.GetUserSession(data.UserId)

	if !isInSession || sessionId != currentSessionId {
		data.SendMessage(data.Trans("session_is_too_old"), true)
		return true
	}

	_, wasInSession := db.LeaveSession(data.UserId)
	data.SubstituteDialog(data.Static.MakeDialogFn("ns", data.UserId, data.Trans, data.Static, nil))
	if wasInSession {
		staticFunctions.UpdateSessionDialogs(sessionId, data.Static)
	}
	return true
}

func (factory *sessionDialogFactory) createVariants(trans i18n.TranslateFunc, sessionId int64) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn() {
			variants = append(variants, dialog.Variant{
				Id:           variant.id,
				Text:         trans(variant.textId),
				RowId:        variant.rowId,
				AdditionalId: strconv.FormatInt(sessionId, 10),
			})
		}
	}
	return
}

func sendSpyfallLocation(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	currentSessionId, isInSession := db.GetUserSession(data.UserId)

	if !isInSession || sessionId != currentSessionId {
		data.SendMessage(data.Trans("no_session_error"), true)
		return true
	}

	isSuccess := staticFunctions.SendSpyfallLocationToAll(data.Static, sessionId)
	if !isSuccess {
		trans := staticFunctions.FindTransFunction(data.UserId, data.Static)
		data.SendMessage(trans("few_players"), true)
	}
	return true
}

func (factory *sessionDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	db := staticFunctions.GetDb(staticData)

	sessionId, isInSession := db.GetUserSession(userId)

	if !isInSession {
		log.Printf("User %d is not in session", userId)
		return nil
	}

	countInSession := db.GetUsersCountInSession(sessionId, false)

	translationMap := map[string]interface{}{
		"Participants": countInSession,
	}

	return &dialog.Dialog{
		Text:     trans("session_title", translationMap),
		Variants: factory.createVariants(trans, sessionId),
	}
}

func (factory *sessionDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	sessionId, _ := strconv.ParseInt(additionalId, 10, 64)
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(sessionId, data)
		}
	}
	return false
}
