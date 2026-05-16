package staticFunctions

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-spy-game-bot/staticData"
	"github.com/nicksnyder/go-i18n/i18n"
)

func normalizeNumberOfReceivers(numberOfReceivers int, userIdsLen int) int {
	if numberOfReceivers < 0 {
		numberOfReceivers = userIdsLen + numberOfReceivers
	}
	if numberOfReceivers < 0 || numberOfReceivers > userIdsLen {
		numberOfReceivers = 0
	}
	return numberOfReceivers
}

func SendThemeToPlayers(staticData *processing.StaticProccessStructs, sessionId int64, userIds []int64, theme string) (err error) {
	db := GetDb(staticData)

	if len(userIds) < 2 {
		return errors.New("few_players")
	}

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)
	if !configCastSuccess {
		return errors.New("Config type is incorrect")
	}

	gameName := db.GetGameName(sessionId)

	gameRules, ok := config.GameRules[gameName]
	if !ok {
		return fmt.Errorf("Game rules for game %s not found", gameName)
	}

	numberOfReceivers := normalizeNumberOfReceivers(gameRules.NumberOfThemeReceivers, len(userIds))
	if numberOfReceivers <= 0 {
		return errors.New("Incorrect number of receivers")
	}

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })

	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)

		var themeMessage string
		if i < numberOfReceivers {
			themeMessage = theme
		} else if gameRules.SpyTheme != "" {
			themeMessage = trans(gameRules.SpyTheme)
		}

		if themeMessage != "" {
			chatId, isFound := db.GetTelegramUserChatId(userId)
			if isFound {
				staticData.Chat.SendMessage(chatId, wrapIntoTelegramSpoiler(themeMessage, trans), 0, true)
			} else {
				db.AddWebMessage(userId, themeMessage, 10)
			}
		}
	}
	return nil
}

func SendThemeToOthers(staticData *processing.StaticProccessStructs, sessionId int64, excludeUserId int64, theme string) (err error) {
	playersInSession := GetDb(staticData).GetUsersInSession(sessionId)
	var playersExceptCurrent []int64
	for _, userId := range playersInSession {
		if userId != excludeUserId {
			playersExceptCurrent = append(playersExceptCurrent, userId)
		}
	}

	return SendThemeToPlayers(staticData, sessionId, playersExceptCurrent, theme)
}

func SendStaticThemeToAll(staticData *processing.StaticProccessStructs, sessionId int64) (err error) {
	db := GetDb(staticData)

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)
	if !configCastSuccess {
		return errors.New("Config type is incorrect")
	}

	gameName := db.GetGameName(sessionId)

	gameRules, ok := config.GameRules[gameName]
	if !ok {
		return fmt.Errorf("Game rules for game %s not found", gameName)
	}

	themesCount := len(gameRules.StaticThemes) + len(gameRules.StaticThemesWithRoles)

	if themesCount == 0 {
		return errors.New("no_static_themes")
	}

	userIds := db.GetUsersInSession(sessionId)

	if len(userIds) < 2 {
		return errors.New("few_players")
	}

	numberOfReceivers := normalizeNumberOfReceivers(gameRules.NumberOfThemeReceivers, len(userIds))
	if numberOfReceivers <= 0 {
		return errors.New("Incorrect number of receivers")
	}

	rand.Shuffle(len(userIds), func(i, j int) {
		userIds[i], userIds[j] = userIds[j], userIds[i]
	})

	themeIdx := rand.Intn(themesCount)

	themeId := "";
	var roles *[]string = nil;
	if themeIdx < len(gameRules.StaticThemes) {
		themeId = gameRules.StaticThemes[themeIdx]
	} else {
		themeId = gameRules.StaticThemesWithRoles[themeIdx - len(gameRules.StaticThemes)].ThemeId
		roles = &gameRules.StaticThemesWithRoles[themeIdx - len(gameRules.StaticThemes)].Roles
	}

	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)

		var theme string
		if i >= numberOfReceivers {
			theme = trans(gameRules.SpyTheme)
		} else if roles != nil && i < len(*roles)-1 {
			theme = trans(gameRules.ThemeTemplate, map[string]interface{}{
				"ThemeId": trans(gameRules.ThemePrefix + themeId),
				"Role":     trans(gameRules.RolePrefix + themeId + "_" + (*roles)[i]),
			})
		} else {
			theme = trans(gameRules.ThemeTemplate, map[string]interface{}{
				"ThemeId": trans(gameRules.ThemePrefix + themeId),
				"Role":     trans(gameRules.RolePrefix + themeId + "_" + (*roles)[len(*roles)-1]),
			})
		}

		chatId, isFound := db.GetTelegramUserChatId(userId)
		if isFound {
			staticData.Chat.SendMessage(chatId, wrapIntoTelegramSpoiler(theme, trans), 0, true)
		} else {
			db.AddWebMessage(userId, theme, 10)
		}
	}

	return nil
}

func getRuneIdx(text []rune, what string) int {
	whatRunes := []rune(what)

	for i := range text {
		found := true
		for j := range whatRunes {
			if text[i+j] != whatRunes[j] {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}

func wrapIntoTelegramSpoiler(text string, trans i18n.TranslateFunc) string {
	// since Telegram spoiler tag shows length, add spaces at the end of each line to obfuscate it
	var finalText strings.Builder
	const maxLineLength = 40
	for {
		strSeparatorPos := strings.IndexRune(text, '\n')
		// this is a quick and ugly fix for the extended unicode mixed into the string
		runeSeparatorPos := getRuneIdx([]rune(text), "\n")
		if runeSeparatorPos == -1 {
			finalText.WriteString(text + "\n")
			break
		}

		// skip long lines
		if runeSeparatorPos > maxLineLength {
			finalText.WriteString(text[:strSeparatorPos] + "\n")
			text = text[strSeparatorPos:]
			continue
		}

		finalText.WriteString(text[:strSeparatorPos] + strings.Repeat(" ", maxLineLength-runeSeparatorPos) + "\n")
		text = text[strSeparatorPos+1:]
	}
	// add an extra static line, since spaces from the last line are cut off
	return "<tg-spoiler>" + finalText.String() + "<i>" + trans("spoiler_terminator") + "</i>" + "</tg-spoiler>"
}

func SendStaticThemeList(data *processing.ProcessData) (err error) {
	db := GetDb(data.Static)

	sessionId, isInSession := db.GetUserSession(data.UserId)
	if !isInSession {
		return errors.New("no_session_title")
	}

	config, configCastSuccess := data.Static.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		return errors.New("Config type is incorrect")
	}

	gameName := db.GetGameName(sessionId)

	gameRules, ok := config.GameRules[gameName]
	if !ok {
		return fmt.Errorf("Game rules for game %s not found", gameName)
	}

	if len(gameRules.StaticThemes) == 0 && len(gameRules.StaticThemesWithRoles) == 0 {
		return errors.New("no_static_themes")
	}

	var themesList []string
	for _, themeId := range gameRules.StaticThemes {
		themesList = append(themesList, data.Trans(gameRules.ThemePrefix + themeId))
	}
	for _, theme := range gameRules.StaticThemesWithRoles {
		themesList = append(themesList, data.Trans(gameRules.ThemePrefix + theme.ThemeId))
	}

	data.SendMessage(strings.Join(themesList[:], "\n"), true)

	return nil
}

func GiveRandomNumbersToPlayers(staticData *processing.StaticProccessStructs, sessionId int64) {
	db := GetDb(staticData)

	userIds := db.GetUsersInSession(sessionId)

	if len(userIds) < 2 {
		return
	}

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })
	for i, userId := range userIds {
		trans := FindTransFunction(userId, staticData)
		theme := trans("player_number_msg", map[string]interface{}{
			"Number": i + 1,
		})

		chatId, isFound := db.GetTelegramUserChatId(userId)
		if isFound {
			staticData.Chat.SendMessage(chatId, theme, 0, true)
		} else {
			db.AddWebMessage(userId, theme, 10)
		}
	}
}
