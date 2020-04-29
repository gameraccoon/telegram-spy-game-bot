package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
)

type sessionVariantPrototype struct {
	id string
	textId string
	process func(int64, *processing.ProcessData) bool
	rowId int
	isActiveFn func() bool
}

type sessionDialogFactory struct {
	variants []sessionVariantPrototype
}

func MakeSessionDialogFactory() dialogFactory.DialogFactory {
	return &(sessionDialogFactory{
		variants: []sessionVariantPrototype{
			sessionVariantPrototype{
				id: "showtheme",
				textId: "show_theme",
				process: testAction,
				isActiveFn: isThemeHidden,
				rowId:1,
			},
			sessionVariantPrototype{
				id: "hidetheme",
				textId: "hide_theme",
				process: testAction,
				isActiveFn: isThemeRevealed,
				rowId:1,
			},
			sessionVariantPrototype{
				id: "discsess",
				textId: "disconnect_session",
				process: disconnectSession,
				rowId:2,
			},
		},
	})
}

func isThemeRevealed() bool {
	return false
}

func isThemeHidden() bool {
	return true
}

func testAction(userId int64, data *processing.ProcessData) bool {
	//data.SubstitudeDialog(data.Static.MakeDialogFn("lc", data.UserId, data.Trans, data.Static))
	return true
}

func disconnectSession(userId int64, data *processing.ProcessData) bool {
	staticFunctions.GetDb(data.Static).DisconnectFromSession(userId)
	data.SubstitudeDialog(data.Static.MakeDialogFn("ns", data.UserId, data.Trans, data.Static))
	return true
}

func (factory *sessionDialogFactory) createVariants(trans i18n.TranslateFunc) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn() {
			variants = append(variants, dialog.Variant{
				Id:   variant.id,
				Text: trans(variant.textId),
				RowId: variant.rowId,
			})
		}
	}
	return
}

func (factory *sessionDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs) *dialog.Dialog {
	db := staticFunctions.GetDb(staticData)

	sessionId, _ := db.GetUserSession(userId)
	countInSession := db.GetUsersCountInSession(sessionId)

	translationMap := map[string]interface{} {
		"SessionId": sessionId,
		"Participants": countInSession,
	}

	return &dialog.Dialog{
		Text:     trans("session_title", translationMap),
		Variants: factory.createVariants(trans),
	}
}

func (factory *sessionDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(data.UserId, data)
		}
	}
	return false
}
