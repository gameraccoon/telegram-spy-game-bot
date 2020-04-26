package staticData

type LanguageData struct {
	Key string
	Name string
}

type StaticConfiguration struct {
	AvailableLanguages []LanguageData
	DefaultLanguage string
	ExtendedLog bool
}
