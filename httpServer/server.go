package httpServer

import (
	"fmt"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-spy-game-bot/database"
	"github.com/gameraccoon/telegram-spy-game-bot/staticFunctions"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type webCaches struct {
	indexHtml           string
	inviteHtml          string
	inviteNoSessionHtml string
	userHtml            string
}

func loadCaches() (caches webCaches, err error) {
	pageHtml, err := os.ReadFile("data/html/index.html")
	if err != nil {
		log.Fatal("Error while reading index.html: ", err)
		return
	}
	caches.indexHtml = string(pageHtml)

	pageHtml, err = os.ReadFile("data/html/invite.html")
	if err != nil {
		log.Fatal("Error while reading invite.html: ", err)
		return
	}
	caches.inviteHtml = string(pageHtml)

	pageHtml, err = os.ReadFile("data/html/invite_no_session.html")
	if err != nil {
		log.Fatal("Error while reading invite_no_session.html: ", err)
		return
	}
	caches.inviteNoSessionHtml = string(pageHtml)

	pageHtml, err = os.ReadFile("data/html/user.html")
	if err != nil {
		log.Fatal("Error while reading user.html: ", err)
		return
	}
	caches.userHtml = string(pageHtml)

	return
}

func servePreloaded(w http.ResponseWriter, page *string) {
	_, err := fmt.Fprint(w, *page)
	if err != nil {
		log.Println("Error serving page: ", err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request, caches *webCaches) {
	servePreloaded(w, &caches.indexHtml)
}

func invitePage(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb, caches *webCaches) {
	gameToken := r.URL.Path[len("/invite/"):]
	if gameToken == "" {
		http.Error(w, "Incorrect URL", http.StatusBadRequest)
		return
	}

	_, isFound := db.GetSessionIdFromToken(gameToken)
	if isFound {
		servePreloaded(w, &caches.inviteHtml)
	} else {
		servePreloaded(w, &caches.inviteNoSessionHtml)
	}
}

func joinGame(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	gameId := r.Form.Get("gameId")

	if gameId == "" {
		http.Error(w, "Incorrect game id, reload the page and try again", http.StatusBadRequest)
		return
	}

	sessionId, isFound := db.GetSessionIdFromToken(gameId)
	if !isFound {
		http.Error(w, "Game not found. Was it ended?", http.StatusBadRequest)
		return
	}

	token := int64(rand.Uint64() & 0x7FFFFFFFFFFFFFFF)

	hasAdded := db.AddWebUser(sessionId, token)

	if !hasAdded {
		http.Error(w, "Can't add new user, try again", http.StatusBadRequest)
		return
	}

	staticFunctions.UpdateSessionDialogs(sessionId, staticData)

	stringToken := strconv.FormatInt(token, 10)

	_, err = w.Write([]byte(stringToken))
	if err != nil {
		return
	}
}

func gamePage(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb, caches *webCaches) {
	if r.Method != "GET" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// get player token from URL
	urlPayload := r.URL.Path[len("/user/"):]
	if urlPayload == "" {
		http.Error(w, "Incorrect URL", http.StatusBadRequest)
		return
	}

	playerTokenStrSplit := strings.Split(urlPayload, "/")
	if len(playerTokenStrSplit) != 2 {
		http.Error(w, "Incorrect URL format", http.StatusBadRequest)
		return
	}

	playerTokenStr := playerTokenStrSplit[1]

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	_, isFound := db.GetWebUserId(playerToken)
	if isFound {
		servePreloaded(w, &caches.userHtml)
	} else {
		servePreloaded(w, &caches.inviteNoSessionHtml)
	}
}

func getLastMessages(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb) {
	if r.Method != "GET" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()

	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	lastMessageIdxStr := r.Form.Get("lastMessageIdx")
	if lastMessageIdxStr == "" {
		http.Error(w, "Incorrect last message index", http.StatusBadRequest)
		return
	}

	lastMessageIdx, err := strconv.Atoi(lastMessageIdxStr)
	if err != nil {
		http.Error(w, "Incorrect last message index", http.StatusBadRequest)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	messages, newLastIdx := db.GetNewRecentWebMessages(userId, lastMessageIdx)

	w.Header().Set("Content-Type", "application/json")
	messagesStr := ""
	for i, message := range messages {
		if i > 0 {
			messagesStr += ","
		}

		sanitizedString := strings.Replace(message, "\n", "<br/>", -1)
		sanitizedString = strings.Replace(sanitizedString, "\"", "\\\"", -1)

		messagesStr += "\"" + sanitizedString + "\""
	}

	playersCount := db.GetUsersCountInSession(sessionId, false)

	_, err = w.Write([]byte("{\"lastMessageIdx\":" + strconv.Itoa(newLastIdx) + ",\"players\":" + strconv.FormatInt(playersCount, 10) + ",\"messages\":[" + messagesStr + "]}"))
}

func sendHiddenMessage(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	message := r.Form.Get("message")
	if message == "" {
		http.Error(w, "The message is empty", http.StatusBadRequest)
		return
	}

	isSucceeded := staticFunctions.SendThemeToOthers(staticData, sessionId, userId, message)

	if !isSucceeded {
		_, err = w.Write([]byte("Not enough players"))
		if err != nil {
			return
		}
	}

	_, err = w.Write([]byte("ok"))
	if err != nil {
		return
	}
}

func sendSpyfallLocation(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	isSucceeded := staticFunctions.SendSpyfallLocationToAll(staticData, sessionId)

	if !isSucceeded {
		_, err = w.Write([]byte("Not enough players"))
		if err != nil {
			return
		}
	}

	_, err = w.Write([]byte("ok"))
}

func leaveGame(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	db.RemoveWebUser(playerToken)

	staticFunctions.UpdateSessionDialogs(sessionId, staticData)

	_, err = w.Write([]byte("ok"))
}

func sendNumbers(w http.ResponseWriter, r *http.Request, db *database.SpyBotDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	staticFunctions.GiveRandomNumbersToPlayers(staticData, sessionId)

	_, err = w.Write([]byte("ok"))
	if err != nil {
		return
	}
}

func HandleHttpRequests(port int, staticData *processing.StaticProccessStructs) {
	db := staticFunctions.GetDb(staticData)

	caches, err := loadCaches()
	if err != nil {
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homePage(w, r, &caches)
	})
	http.HandleFunc("/invite/", func(w http.ResponseWriter, r *http.Request) {
		invitePage(w, r, db, &caches)
	})
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		joinGame(w, r, db, staticData)
	})
	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		gamePage(w, r, db, &caches)
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		getLastMessages(w, r, db)
	})
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		sendHiddenMessage(w, r, db, staticData)
	})
	http.HandleFunc("/spyfall", func(w http.ResponseWriter, r *http.Request) {
		sendSpyfallLocation(w, r, db, staticData)
	})
	http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
		leaveGame(w, r, db, staticData)
	})
	http.HandleFunc("/numbers", func(w http.ResponseWriter, r *http.Request) {
		sendNumbers(w, r, db, staticData)
	})

	addr := ":" + strconv.Itoa(port)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
