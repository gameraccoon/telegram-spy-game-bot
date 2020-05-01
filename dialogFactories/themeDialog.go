package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"log"
)

type themeVariantPrototype struct {
	id string
	textId string
	process func(*processing.ProcessData) bool
	rowId int
	isActiveFn func(bool) bool
}

type themeDialogFactory struct {
	variants []themeVariantPrototype
}

func MakeThemeDialogFactory() dialogFactory.DialogFactory {
	return &(themeDialogFactory{
		variants: []themeVariantPrototype{
			themeVariantPrototype{
				id: "showtheme",
				textId: "show_theme",
				process: revealTheme,
				isActiveFn: isThemeHidden,
				rowId:1,
			},
			themeVariantPrototype{
				id: "hidetheme",
				textId: "hide_theme",
				process: hideTheme,
				isActiveFn: isThemeRevealed,
				rowId:1,
			},
		},
	})
}

func isThemeHidden(isRevealed bool) bool {
	return !isRevealed
}

func isThemeRevealed(isRevealed bool) bool {
	return isRevealed
}

func revealTheme(data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetThemeRevealed(data.UserId, true)
	theme := db.GetUserTheme(data.UserId)
	data.SubstitudeDialog(data.Static.MakeDialogFn("th", data.UserId, data.Trans, data.Static, theme))
	return true
}

func hideTheme(data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetThemeRevealed(data.UserId, false)
	data.SubstitudeDialog(data.Static.MakeDialogFn("th", data.UserId, data.Trans, data.Static, ""))
	return true
}

func (factory *themeDialogFactory) createVariants(trans i18n.TranslateFunc, isRevealed bool) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn(isRevealed) {
			variants = append(variants, dialog.Variant{
				Id:   variant.id,
				Text: trans(variant.textId),
				RowId: variant.rowId,
			})
		}
	}
	return
}

func (factory *themeDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	db := staticFunctions.GetDb(staticData)

	isRevealed := db.IsThemeRevealed(userId)

	message, ok := customData.(string)
	if !ok {
		log.Print("message is not set for theme dialog")
	}

	var title string
	if isRevealed {
		title = message
	} else {
		title = trans("theme_hidden")
	}

	return &dialog.Dialog{
		Text:     title,
		Variants: factory.createVariants(trans, isRevealed),
	}
}

func (factory *themeDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(data)
		}
	}
	return false
}
