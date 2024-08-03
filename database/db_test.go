package database

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	testDbPath = "./testDb.db"
)

func dropDatabase(fileName string) {
	_ = os.Remove(fileName)
}

func clearDb() {
	dropDatabase(testDbPath)
}

func connectDb(t *testing.T) *SpyBotDb {
	assert := require.New(t)
	db, err := ConnectDb(testDbPath)

	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return nil
	}
	return db
}

func createDbAndConnect(t *testing.T) *SpyBotDb {
	clearDb()
	return connectDb(t)
}

func TestConnection(t *testing.T) {
	assert := require.New(t)
	dropDatabase(testDbPath)

	db, err := ConnectDb(testDbPath)

	defer dropDatabase(testDbPath)
	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return
	}

	assert.True(db.IsConnectionOpened())

	db.Disconnect()

	assert.False(db.IsConnectionOpened())
}

func TestSanitizeString(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	testText := "text'test''test\"test\\"

	db.SetDatabaseVersion(testText)
	assert.Equal(testText, db.GetDatabaseVersion())
}

func TestDatabaseVersion(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}

	{
		version := db.GetDatabaseVersion()
		assert.Equal(latestVersion, version)
	}

	{
		db.SetDatabaseVersion("1.0")
		version := db.GetDatabaseVersion()
		assert.Equal("1.0", version)
	}

	db.Disconnect()

	{
		db = connectDb(t)
		version := db.GetDatabaseVersion()
		assert.Equal("1.0", version)
		db.Disconnect()
	}

	{
		db = connectDb(t)
		db.SetDatabaseVersion("1.2")
		db.Disconnect()
	}

	{
		db = connectDb(t)
		version := db.GetDatabaseVersion()
		assert.Equal("1.2", version)
		db.Disconnect()
	}
}

func TestGetUserId(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	var chatId1 int64 = 321
	var chatId2 int64 = 123

	id1 := db.GetOrCreateTelegramUserId(chatId1, "")
	id2 := db.GetOrCreateTelegramUserId(chatId1, "")
	id3 := db.GetOrCreateTelegramUserId(chatId2, "")

	assert.Equal(id1, id2)
	assert.NotEqual(id1, id3)

	userChatId1, found := db.GetTelegramUserChatId(id1)
	assert.True(found)
	assert.Equal(chatId1, userChatId1)
	userChatId3, found := db.GetTelegramUserChatId(id3)
	assert.True(found)
	assert.Equal(chatId2, userChatId3)
}

func TestUserLanguage(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "")
	userId2 := db.GetOrCreateTelegramUserId(321, "")

	db.SetUserLanguage(userId1, "en-US")

	{
		lang1 := db.GetUserLanguage(userId1)
		lang2 := db.GetUserLanguage(userId2)
		assert.Equal("en-US", lang1)
		assert.Equal("", lang2)
	}

	// in case of some side effects
	{
		lang1 := db.GetUserLanguage(userId1)
		lang2 := db.GetUserLanguage(userId2)
		assert.Equal("en-US", lang1)
		assert.Equal("", lang2)
	}
}

func TestUserSession(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "")
	userId2 := db.GetOrCreateTelegramUserId(321, "")

	sessionId, _, _ := db.CreateSession(userId1)
	assert.True(db.DoesSessionExist(sessionId))

	{
		token, isFound1 := db.GetTokenFromSessionId(sessionId)
		newSessionId, isFound2 := db.GetSessionIdFromToken(token)
		assert.True(isFound1)
		assert.True(isFound2)
		assert.Equal(sessionId, newSessionId)
	}

	{
		sessionId1, isInSession1 := db.GetUserSession(userId1)
		_, isInSession2 := db.GetUserSession(userId2)
		assert.True(isInSession1)
		assert.False(isInSession2)
		assert.Equal(sessionId, sessionId1)
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId1, true))
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId1, false))

		users := db.GetUsersInSession(sessionId)
		assert.Equal(1, len(users))
		if len(users) > 0 {
			assert.Equal(userId1, users[0])
		}
	}

	db.ConnectToSession(userId2, sessionId)

	{
		sessionId1, isInSession1 := db.GetUserSession(userId1)
		sessionId2, isInSession2 := db.GetUserSession(userId2)
		assert.True(isInSession1)
		assert.True(isInSession2)
		assert.Equal(sessionId, sessionId1)
		assert.Equal(sessionId, sessionId2)
		assert.Equal(int64(2), db.GetUsersCountInSession(sessionId, true))
		assert.Equal(int64(2), db.GetUsersCountInSession(sessionId, false))
	}

	db.LeaveSession(userId1)
	assert.True(db.DoesSessionExist(sessionId))

	{
		_, isInSession1 := db.GetUserSession(userId1)
		sessionId2, isInSession2 := db.GetUserSession(userId2)
		assert.False(isInSession1)
		assert.True(isInSession2)
		assert.Equal(sessionId, sessionId2)
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, true))
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, false))
	}

	db.LeaveSession(userId2)
	assert.False(db.DoesSessionExist(sessionId))

	{
		_, isInSession1 := db.GetUserSession(userId1)
		_, isInSession2 := db.GetUserSession(userId2)
		assert.False(isInSession1)
		assert.False(isInSession2)
		assert.Equal(int64(0), db.GetUsersCountInSession(sessionId, true))
		assert.Equal(int64(0), db.GetUsersCountInSession(sessionId, false))
	}
}

func TestSessionMessageId(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "")
	sessionMessageId := int64(32)

	{
		_, isFound := db.GetSessionMessageId(userId1)
		assert.False(isFound)
	}
	db.SetSessionMessageId(userId1, sessionMessageId)

	{
		sessionId, isFound := db.GetSessionMessageId(userId1)
		assert.True(isFound)
		assert.Equal(sessionMessageId, sessionId)
	}
}

func TestAddWebUser(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	webUserToken := int64(10)

	// we can add web users only if we have a session
	userId := db.GetOrCreateTelegramUserId(123, "")
	sessionId, _, _ := db.CreateSession(userId)

	assert.False(db.DoesWebUserExist(webUserToken))

	wasAdded := db.AddWebUser(sessionId, webUserToken)
	assert.True(wasAdded)

	assert.True(db.DoesWebUserExist(webUserToken))

	wasAdded = db.AddWebUser(sessionId, webUserToken)
	assert.False(wasAdded) // same token

	assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, true))
	assert.Equal(int64(2), db.GetUsersCountInSession(sessionId, false))

	users := db.GetUsersInSession(sessionId)
	assert.Equal(2, len(users))

	for _, user := range users {
		if user == userId {
			continue
		}
		userSessionId, isInSession := db.GetUserSession(user)
		assert.True(isInSession)
		assert.Equal(sessionId, userSessionId)
	}

	_, isFound := db.GetWebUserId(webUserToken)
	assert.True(isFound)

	sessionToken, _ := db.GetTokenFromSessionId(sessionId)

	// web users are not counted for the session survival
	db.LeaveSession(userId)

	assert.False(db.DoesSessionExist(sessionId))
	_, isSessionFound := db.GetSessionIdFromToken(sessionToken)
	assert.False(isSessionFound)
	assert.False(db.DoesWebUserExist(webUserToken))
	_, isFound = db.GetWebUserId(webUserToken)
	assert.False(isFound)
}

func TestRemoveWebUser(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	webUserToken := int64(10)

	userId := db.GetOrCreateTelegramUserId(123, "")
	sessionId, _, _ := db.CreateSession(userId)

	db.AddWebUser(sessionId, webUserToken)

	assert.True(db.DoesWebUserExist(webUserToken))

	db.RemoveWebUser(webUserToken)

	assert.False(db.DoesWebUserExist(webUserToken))

	assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, false))
}

func TestWebMessages(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId := db.GetOrCreateTelegramUserId(123, "")
	sessionId, _, _ := db.CreateSession(userId)

	webUserToken := int64(42)
	db.AddWebUser(sessionId, webUserToken)
	webUserId, _ := db.GetWebUserId(webUserToken)

	{
		commands, newLastIndex := db.GetNewRecentWebMessages(webUserId, -1)
		assert.Equal(0, len(commands))
		assert.Equal(-1, newLastIndex)
	}

	db.AddWebMessage(webUserId, "command1", 10)
	db.AddWebMessage(webUserId, "command2", 10)
	db.AddWebMessage(webUserId, "command3", 10)

	{
		commands, newLastIndex := db.GetNewRecentWebMessages(webUserId, -1)
		assert.Equal(3, len(commands))
		assert.Equal(2, newLastIndex)
		assert.Equal("command1", commands[0])
		assert.Equal("command2", commands[1])
		assert.Equal("command3", commands[2])
	}

	{
		commands, newLastIndex := db.GetNewRecentWebMessages(webUserId, 0)
		assert.Equal(2, len(commands))
		assert.Equal(2, newLastIndex)
		assert.Equal("command2", commands[0])
		assert.Equal("command3", commands[1])
	}

	{
		commands, newLastIndex := db.GetNewRecentWebMessages(webUserId, 1)
		assert.Equal(1, len(commands))
		assert.Equal(2, newLastIndex)
		assert.Equal("command3", commands[0])
	}

	{
		commands, newLastIndex := db.GetNewRecentWebMessages(webUserId, 2)
		assert.Equal(0, len(commands))
		assert.Equal(2, newLastIndex)
	}

	db.AddWebMessage(webUserId, "command4", 2)

	{
		messages, newLastIndex := db.GetNewRecentWebMessages(webUserId, 0)
		assert.Equal(2, len(messages))
		assert.Equal(3, newLastIndex)
		assert.Equal("command3", messages[0])
		assert.Equal("command4", messages[1])
	}

	db.LeaveSession(userId)

	{
		messages, newLastIndex := db.GetNewRecentWebMessages(webUserId, 0)
		assert.Equal(0, len(messages))
		assert.Equal(0, newLastIndex)
	}
}

// regression test
func TestWebMessagesClearing(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId := db.GetOrCreateTelegramUserId(123, "")

	{
		sessionId, _, _ := db.CreateSession(userId)

		webUserToken := int64(42)
		db.AddWebUser(sessionId, webUserToken)
		webUserId, _ := db.GetWebUserId(webUserToken)

		db.AddWebMessage(webUserId, "command1", 10)

		db.RemoveWebUser(webUserToken)
	}

	{
		sessionId, _, _ := db.CreateSession(userId)

		webUserToken := int64(63)
		db.AddWebUser(sessionId, webUserToken)
		webUserId, _ := db.GetWebUserId(webUserToken)

		commands, newLastIndex := db.GetNewRecentWebMessages(webUserId, -1)
		assert.Equal(0, len(commands))
		assert.Equal(-1, newLastIndex)
	}
}
