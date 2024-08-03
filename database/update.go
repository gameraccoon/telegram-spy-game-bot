package database

import (
	"fmt"
	"log"
)

const (
	minimalVersion = "0.1"
	latestVersion  = "0.2"
)

type dbUpdater struct {
	version  string
	updateDb func(db *SpyBotDb)
}

func UpdateVersion(db *SpyBotDb) {
	currentVersion := db.GetDatabaseVersion()

	if currentVersion != latestVersion {
		updaters := makeUpdaters(currentVersion, latestVersion)

		log.Printf("Update DB version from %s to %s in %d iterations", currentVersion, latestVersion, len(updaters))
		for _, updater := range updaters {
			log.Printf("Updating to %s", updater.version)
			updater.updateDb(db)
		}
	}

	db.SetDatabaseVersion(latestVersion)
}

func makeUpdaters(versionFrom string, versionTo string) (updaters []dbUpdater) {
	allUpdaters := makeAllUpdaters()

	isFirstFound := (versionFrom == minimalVersion)
	for _, updater := range allUpdaters {
		if isFirstFound {
			updaters = append(updaters, updater)
			if updater.version == versionTo {
				break
			}
		} else {
			if updater.version == versionFrom {
				isFirstFound = true
			}
		}
	}

	if len(updaters) > 0 {
		lastFoundVersion := updaters[len(updaters)-1].version
		if lastFoundVersion != versionTo {
			log.Fatalf("Last version updater not found. Expected: %s Found: %s", versionTo, lastFoundVersion)
		}
	}
	return
}

func makeAllUpdaters() []dbUpdater {
	return []dbUpdater{
		{
			version: "0.2",
			updateDb: func(db *SpyBotDb) {
				// for each 'users' record create a new record in the 'telegram_users' table
				rows, err := db.db.Query("SELECT id, chat_id, language, IFNULL(current_session_message, 0) FROM users")
				if err != nil {
					log.Fatalf("Error while selecting users: %s", err)
				}
				defer func() {
					err := rows.Close()
					if err != nil {
						log.Fatalf("Error while closing rows: %s", err)
					}
				}()

				dataToTransfer := make([][]interface{}, 0)
				for rows.Next() {
					var id int64
					var chatId int64
					var language string
					var currentSessionMessage int64
					err := rows.Scan(&id, &chatId, &language, &currentSessionMessage)
					if err != nil {
						log.Fatalf("Error while scanning id: %s", err)
					}

					dataToTransfer = append(dataToTransfer, []interface{}{id, chatId, language, currentSessionMessage})
				}

				err = rows.Close()
				if err != nil {
					log.Fatalf("Error while closing rows: %s", err)
				}

				for _, data := range dataToTransfer {
					currentSessionMessage := fmt.Sprintf("%d", data[3])
					if currentSessionMessage == "0" {
						currentSessionMessage = "NULL"
					}
					db.db.Exec(fmt.Sprintf("INSERT INTO telegram_users (user_id, chat_id, language, current_session_message) VALUES (%d, %d, '%s', %s)", data[0], data[1], data[2], currentSessionMessage))
				}

				// remove unused columns from 'users' table
				db.db.Exec("ALTER TABLE users RENAME TO users_old")
				db.db.Exec("CREATE TABLE" +
					" users(id INTEGER NOT NULL PRIMARY KEY" +
					",current_session INTEGER" +
					")")
				db.db.Exec("INSERT INTO users (id, current_session) SELECT id, current_session FROM users_old")
				db.db.Exec("DROP TABLE users_old")
			},
		},
	}
}
