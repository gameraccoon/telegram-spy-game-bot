package dialogFactories

import (
	"fmt"
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"log"
	"strconv"
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
				id: "share",
				textId: "share_link",
				rowId:1,
			},
			sessionVariantPrototype{
				id: "discsess",
				textId: "disconnect_session",
				process: disconnectSession,
				rowId:1,
			},
			/*sessionVariantPrototype{
				id: "showtheme",
				textId: "show_theme",
				process: testAction,
				isActiveFn: isThemeHidden,
				rowId:2,
			},
			sessionVariantPrototype{
				id: "hidetheme",
				textId: "hide_theme",
				process: testAction,
				isActiveFn: isThemeRevealed,
				rowId:2,
			},*/
		},
	})
}

func disconnectSession(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	currentSessionId, isInSession := db.GetUserSession(data.UserId)

	if !isInSession || sessionId != currentSessionId {
		data.SendMessage(data.Trans("session_is_too_old"))
		return true
	}

	_, wasInSession := db.LeaveSession(data.UserId)
	data.SubstitudeDialog(data.Static.MakeDialogFn("ns", data.UserId, data.Trans, data.Static, nil))
	if wasInSession {
		staticFunctions.UpdateSessionDialogs(sessionId, data.Static)
	}
	return true
}

func (factory *sessionDialogFactory) createVariants(trans i18n.TranslateFunc, sessionId int64, sessionUrl string) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn() {
			var url string
			if variant.process == nil {
				url = sessionUrl
			}

			variants = append(variants, dialog.Variant{
				Id:   variant.id,
				Text: trans(variant.textId),
				Url: url,
				RowId: variant.rowId,
				AdditionalId: strconv.FormatInt(sessionId, 10),
			})
		}
	}
	return
}

func (factory *sessionDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	db := staticFunctions.GetDb(staticData)

	sessionId, _ := db.GetUserSession(userId)
	countInSession := db.GetUsersCountInSession(sessionId)

	translationMap := map[string]interface{} {
		"Participants": countInSession,
	}

	sessionToken, isFound := db.GetTokenFromSessionId(sessionId)

	if !isFound {
		log.Printf("Can't find session token for sessionId %d", sessionId)
	}

	sessionUrl := fmt.Sprintf(
		"https://telegram.me/share/url?url=https://t.me/%s?start=%s",
		staticData.BotName,
		sessionToken,
	)

	return &dialog.Dialog{
		Text:     trans("session_title", translationMap),
		Variants: factory.createVariants(trans, sessionId, sessionUrl),
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
