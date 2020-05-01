package database

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	dbBase "github.com/gameraccoon/telegram-bot-skeleton/database"
	"log"
	"sync"
)

type SpyBotDb struct {
	db dbBase.Database
	mutex sync.Mutex
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func ConnectDb(path string) (database *SpyBotDb, err error) {
	database = &SpyBotDb{}

	err = database.db.Connect(path)

	if err != nil {
		return
	}

	database.db.Exec("PRAGMA foreign_keys = ON")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" global_vars(name TEXT PRIMARY KEY" +
		",integer_value INTEGER" +
		",string_value TEXT" +
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" sessions(id INTEGER NOT NULL PRIMARY KEY" +
		",token TEXT NOT NULL" +
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" users(id INTEGER NOT NULL PRIMARY KEY" +
		",chat_id INTEGER UNIQUE NOT NULL" +
		",language TEXT NOT NULL" +

		// session related data
		",is_ready INTEGER NOT NULL" +
		",current_session INTEGER" +
		",current_session_message INTEGER" +
		",FOREIGN KEY(current_session) REFERENCES sessions(id) ON DELETE SET NULL" +
		")")

	database.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" chat_id_index ON users(chat_id)")

	database.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" token_index ON sessions(token)")

	database.db.Exec("CREATE INDEX IF NOT EXISTS" +
		" current_session_index ON users(current_session)")

	return
}

func (database *SpyBotDb) IsConnectionOpened() bool {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	return database.db.IsConnectionOpened()
}

func (database *SpyBotDb) Disconnect() {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Disconnect()
}

func (database *SpyBotDb) GetDatabaseVersion() (version string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query("SELECT string_value FROM global_vars WHERE name='version'")

	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		// that means it's a new clean database
		version = latestVersion
	}

	return
}

func (database *SpyBotDb) SetDatabaseVersion(version string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec("DELETE FROM global_vars WHERE name='version'")

	safeVersion := dbBase.SanitizeString(version)
	database.db.Exec(fmt.Sprintf("INSERT INTO global_vars (name, string_value) VALUES ('version', '%s')", safeVersion))
}

func (database *SpyBotDb) GetUserId(chatId int64, userLangCode string) (userId int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("INSERT OR IGNORE INTO users(chat_id, language, is_ready) "+
		"VALUES (%d, '%s', 0)", chatId, userLangCode))

	rows, err := database.db.Query(fmt.Sprintf("SELECT id FROM users WHERE chat_id=%d", chatId))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&userId)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal("No user found")
	}

	return
}

func (database *SpyBotDb) GetChatId(userId int64) (chatId int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT chat_id FROM users WHERE id=%d", userId))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&chatId)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal("No user found")
	}

	return
}

func (database *SpyBotDb) GetUserChatId(userId int64) (chatId int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT chat_id FROM users WHERE id=%d", userId))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&chatId)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal("No user found")
	}

	return
}

func (database *SpyBotDb) getLastInsertedItemId() (id int64) {
	rows, err := database.db.Query("SELECT last_insert_rowid()")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal("No item found")
	}
	return -1
}

func (database *SpyBotDb) SetUserLanguage(userId int64, language string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET language='%s' WHERE id=%d", language, userId))
}

func (database *SpyBotDb) GetUserLanguage(userId int64) (language string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT language FROM users WHERE id=%d AND language IS NOT NULL", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&language)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		// empty language
	}

	return
}

func (database *SpyBotDb) GetUserSession(userId int64) (sessionId int64, isInSession bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT current_session FROM users WHERE id=%d AND current_session IS NOT NULL", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&sessionId)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			isInSession = true
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *SpyBotDb) DoesSessionExist(sessionId int64) (isExists bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT 1 FROM sessions WHERE id=%d LIMIT 1", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	isExists = rows.Next();

	return
}

func (database *SpyBotDb) CreateSession(userId int64) (sessionId int64, previousSessionId int64, wasInSession bool) {
	previousSessionId, wasInSession = database.DisconnectFromSession(userId)

	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec("INSERT INTO sessions (token) VALUES (strftime('%s', 'now') || '-' || abs(random() % 100000))")

	sessionId = database.getLastInsertedItemId()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session=%d WHERE id=%d", sessionId, userId))

	return
}

func (database *SpyBotDb) ConnectToSession(userId int64, sessionId int64) (isSucceeded bool, previousSessionId int64, wasInSession bool) {
	if !database.DoesSessionExist(sessionId) {
		return
	}

	previousSessionId, wasInSession = database.DisconnectFromSession(userId)

	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session=%d WHERE id=%d", sessionId, userId))

	isSucceeded = true
	return
}

func (database *SpyBotDb) GetUsersCountInSession(sessionId int64) (usersCount int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT COUNT(*) FROM users WHERE current_session=%d", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&usersCount)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *SpyBotDb) GetUsersInSession(sessionId int64) (users []int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT id FROM users WHERE current_session=%d", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var userId int64
		err := rows.Scan(&userId)
		if err != nil {
			log.Fatal(err.Error())
		}
		users = append(users, userId)
	}

	return
}

func (database *SpyBotDb) DisconnectFromSession(userId int64) (sessionId int64, wasInSession bool) {
	sessionId, wasInSession = database.GetUserSession(userId)

	if !wasInSession {
		return
	}

	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session=NULL WHERE id=%d", userId))

	// delete session if it's empty
	database.mutex.Unlock()
	if database.GetUsersCountInSession(sessionId) == 0 {
		database.mutex.Lock()
		database.db.Exec(fmt.Sprintf("DELETE FROM sessions WHERE id=%d", sessionId))
		database.mutex.Unlock()
	}
	database.mutex.Lock()

	return
}

func (database *SpyBotDb) SetSessionMessageId(userId int64, messageId int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session_message=%d WHERE id=%d", messageId, userId))
}

func (database *SpyBotDb) GetSessionMessageId(userId int64) (messageId int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT current_session_message FROM users WHERE id=%d AND current_session_message IS NOT NULL", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&messageId)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *SpyBotDb) GetSessionIdFromToken(token string) (sessionId int64, isFound bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT id FROM sessions WHERE token='%s' LIMIT 1", dbBase.SanitizeString(token)))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&sessionId)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			isFound = true
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *SpyBotDb) GetTokenFromSessionId(sessionId int64) (token string, isFound bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT token FROM sessions WHERE id=%d LIMIT 1", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&token)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			isFound = true
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}
