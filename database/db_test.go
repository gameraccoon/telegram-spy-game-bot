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
	os.Remove(fileName)
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

	id1 := db.GetUserId(chatId1, "")
	id2 := db.GetUserId(chatId1, "")
	id3 := db.GetUserId(chatId2, "")

	assert.Equal(id1, id2)
	assert.NotEqual(id1, id3)

	assert.Equal(chatId1, db.GetUserChatId(id1))
	assert.Equal(chatId2, db.GetUserChatId(id3))
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

	userId1 := db.GetUserId(123, "")
	userId2 := db.GetUserId(321, "")

	db.SetUserLanguage(userId1, "en-US")

	{
		lang1 := db.GetUserLanguage(userId1)
		lang2 := db.GetUserLanguage(userId2)
		assert.Equal("en-US", lang1)
		assert.Equal("", lang2)
	}

	// in case of some side-effects
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

	userId1 := db.GetUserId(123, "")
	userId2 := db.GetUserId(321, "")

	db.CreateSession(userId1)

	{
		sessionId1, isInSession1 := db.GetUserSession(userId1)
		_, isInSession2 := db.GetUserSession(userId2)
		assert.True(isInSession1)
		assert.False(isInSession2)
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId1))
	}

	sessionId, _ := db.GetUserSession(userId1)
	db.ConnectToSession(userId2, sessionId)

	{
		sessionId1, isInSession1 := db.GetUserSession(userId1)
		sessionId2, isInSession2 := db.GetUserSession(userId2)
		assert.True(isInSession1)
		assert.True(isInSession2)
		assert.Equal(sessionId, sessionId1)
		assert.Equal(sessionId, sessionId2)
		assert.Equal(int64(2), db.GetUsersCountInSession(sessionId))
	}

	db.DisconnectFromSession(userId1)

	{
		_, isInSession1 := db.GetUserSession(userId1)
		sessionId2, isInSession2 := db.GetUserSession(userId2)
		assert.False(isInSession1)
		assert.True(isInSession2)
		assert.Equal(sessionId, sessionId2)
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId))
	}

	db.DisconnectFromSession(userId2)

	{
		_, isInSession1 := db.GetUserSession(userId1)
		_, isInSession2 := db.GetUserSession(userId2)
		assert.False(isInSession1)
		assert.False(isInSession2)
		assert.Equal(int64(0), db.GetUsersCountInSession(sessionId))
	}
}
