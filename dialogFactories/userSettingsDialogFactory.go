package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-spy-game-bot/staticData"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"github.com/nicksnyder/go-i18n/i18n"
)

type userSettingsData struct {
	userId     int64
	staticData *processing.StaticProccessStructs
}

type userSettingsVariantPrototype struct {
	id         string
	textId     string
	process    func(int64, *processing.ProcessData) bool
	rowId      int
	isActiveFn func(*userSettingsData) bool
}

type userSettingsDialogFactory struct {
	variants []userSettingsVariantPrototype
}

func MakeUserSettingsDialogFactory() dialogFactory.DialogFactory {
	return &(userSettingsDialogFactory{
		variants: []userSettingsVariantPrototype{
			userSettingsVariantPrototype{
				id:      "lang",
				textId:  "change_language",
				process: changeLanguage,
				rowId:   1,
			},
		},
	})
}

func changeLanguage(userId int64, data *processing.ProcessData) bool {
	data.SubstituteDialog(data.Static.MakeDialogFn("lc", data.UserId, data.Trans, data.Static, nil))
	return true
}

func (factory *userSettingsDialogFactory) createVariants(settingsData *userSettingsData, trans i18n.TranslateFunc) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn(settingsData) {
			variants = append(variants, dialog.Variant{
				Id:    variant.id,
				Text:  trans(variant.textId),
				RowId: variant.rowId,
			})
		}
	}
	return
}

func (factory *userSettingsDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	settingsData := userSettingsData{
		userId:     userId,
		staticData: staticData,
	}

	db := staticFunctions.GetDb(staticData)

	language := db.GetUserLanguage(userId)

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	langName := language

	for _, lang := range config.AvailableLanguages {
		if lang.Key == language {
			langName = lang.Name
			break
		}
	}

	translationMap := map[string]interface{}{
		"Lang": langName,
	}

	return &dialog.Dialog{
		Text:     trans("user_settings_title", translationMap),
		Variants: factory.createVariants(&settingsData, trans),
	}
}

func (factory *userSettingsDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(data.UserId, data)
		}
	}
	return false
}
