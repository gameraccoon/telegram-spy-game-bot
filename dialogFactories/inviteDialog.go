package dialogFactories

import (
	"fmt"
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-spy-game-bot/staticData"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"github.com/nicksnyder/go-i18n/i18n"
	"log"
	"strconv"
)

type inviteVariantPrototype struct {
	id         string
	textId     string
	process    func(int64, *processing.ProcessData) bool
	rowId      int
	isActiveFn func() bool
}

type inviteDialogFactory struct {
	variants []inviteVariantPrototype
}

func MakeInviteDialogFactory() dialogFactory.DialogFactory {
	return &(inviteDialogFactory{
		variants: []inviteVariantPrototype{
			inviteVariantPrototype{
				id:      "artist",
				textId:  "invite_artist",
				process: shareArtistLink,
				rowId:   1,
			},
			inviteVariantPrototype{
				id:      "spyfall",
				textId:  "invite_spyfall",
				process: shareSpyfallLink,
				rowId:   2,
			},
		},
	})
}

func shareArtistLink(sessionId int64, data *processing.ProcessData) bool {
	return shareLink(sessionId, "fake-artist", data)
}

func shareSpyfallLink(sessionId int64, data *processing.ProcessData) bool {
	return shareLink(sessionId, "spyfall", data)
}

func shareLink(sessionId int64, gameType string, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	staticData := data.Static

	currentSessionId, isInSession := db.GetUserSession(data.UserId)

	if !isInSession || sessionId != currentSessionId {
		data.SendMessage(data.Trans("session_is_too_old"), true)
		return true
	}

	sessionToken, isFound := db.GetTokenFromSessionId(sessionId)

	if !isFound {
		log.Printf("Can't find session token for sessionId %d", sessionId)
	}

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	data.SendMessage("Share this link with your friends to invite them to the game:", true)

	data.SendMessage(fmt.Sprintf(
		"Link to join the game:\n%s/invite/%s/%s",
		config.ShareWebAddress,
		gameType,
		sessionToken,
	),
		true)

	data.SendMessage(fmt.Sprintf(
		"Or show them this QR code:\nhttps://api.qrserver.com/v1/create-qr-code/?size=150x150&margin=10&data=%s/invite/%s/%s",
		config.ShareWebAddress,
		gameType,
		sessionToken,
	),
		false)

	return true
}

func (factory *inviteDialogFactory) createVariants(trans i18n.TranslateFunc, sessionId int64) (variants []dialog.Variant) {
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

func (factory *inviteDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	db := staticFunctions.GetDb(staticData)

	sessionId, isInSession := db.GetUserSession(userId)

	if !isInSession {
		log.Printf("User %d is not in session", userId)
		return nil
	}

	return &dialog.Dialog{
		Text:     trans("invite_title"),
		Variants: factory.createVariants(trans, sessionId),
	}
}

func (factory *inviteDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	sessionId, _ := strconv.ParseInt(additionalId, 10, 64)
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(sessionId, data)
		}
	}
	return false
}
