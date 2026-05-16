package staticData

type LanguageData struct {
	Key  string
	Name string
}

type StaticThemeWithRoles struct {
	ThemeId string
	Roles []string
}

type GameRules struct {
	StaticThemes []string
	StaticThemesWithRoles []StaticThemeWithRoles
	NumberOfThemeReceivers int // negative to send to "all but N", e.g. -1 to send to all but 1
	SpyTheme string // empty to not send spy theme

	ThemeTemplate string
	ThemePrefix string
	RolePrefix string
}

type StaticConfiguration struct {
	AvailableLanguages []LanguageData
	DefaultLanguage string
	ExtendedLog bool
	GameRules map[string]GameRules
	RunHttpServer bool
	HttpServerPort int
	ShareWebAddress string
}
