package staticData

type LanguageData struct {
	Key string
	Name string
}

type SpyfallLocation struct {
	LocationId string
	Roles []string
}

type StaticConfiguration struct {
	AvailableLanguages []LanguageData
	DefaultLanguage string
	ExtendedLog bool
	SpyfallLocations []SpyfallLocation
}
