package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
)

type noSessionVariantPrototype struct {
	id string
	textId string
	process func(*processing.ProcessData) bool
	rowId int
	isActiveFn func() bool
}

type noSessionDialogFactory struct {
	variants []noSessionVariantPrototype
}

func MakeNoSessionDialogFactory() dialogFactory.DialogFactory {
	return &(noSessionDialogFactory{
		variants: []noSessionVariantPrototype{
			/*noSessionVariantPrototype{
				id: "connsess",
				textId: "connect_to_session",
				process: connectToSession,
				rowId:2,
			},*/
			noSessionVariantPrototype{
				id: "createsess",
				textId: "create_session",
				process: createNewSession,
				rowId:2,
			},
		},
	})
}

func connectToSession(data *processing.ProcessData) bool {
	data.Static.SetUserStateTextProcessor(data.UserId, &processing.AwaitingTextProcessorData{
		ProcessorId: "connectSession",
	})
	data.SendMessage(data.Trans("send_session_id"))
	return true
}

func createNewSession(data *processing.ProcessData) bool {
	_, previousSessionId, wasInSession := staticFunctions.GetDb(data.Static).CreateSession(data.UserId)
	staticFunctions.SendSessionDialog(data)
	if wasInSession {
		staticFunctions.UpdateSessionDialogs(previousSessionId, data.Static)
	}
	return true
}

func (factory *noSessionDialogFactory) createVariants(trans i18n.TranslateFunc) (variants []dialog.Variant) {
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

func (factory *noSessionDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	return &dialog.Dialog{
		Text:     trans("no_session_title"),
		Variants: factory.createVariants(trans),
	}
}

func (factory *noSessionDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(data)
		}
	}
	return false
}
