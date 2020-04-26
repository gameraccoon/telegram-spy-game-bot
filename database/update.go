package database

import (
	"log"
)

const (
	minimalVersion = "0.1"
	latestVersion  = "0.1"
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
		lastFoundVersion := updaters[len(updaters) - 1].version
		if lastFoundVersion != versionTo {
			log.Fatalf("Last version updater not found. Expected: %s Found: %s", versionTo, lastFoundVersion)
		}
	}
	return
}

func makeAllUpdaters() []dbUpdater {
	return []dbUpdater{
		// dbUpdater{
		// 	version: "0.2",
		// 	updateDb: func(db *SpyBotDb) {
		// 		db.db.Exec("ALTER TABLE wallets ADD COLUMN contract_address TEXT NOT NULL DEFAULT('')")
		// 	},
		// },
	}
}
