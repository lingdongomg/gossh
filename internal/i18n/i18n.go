package i18n

import (
	"sync"
)

// Language represents a supported language
type Language string

const (
	LangEN Language = "en"
	LangZH Language = "zh"
)

var (
	currentLang Language = LangEN
	mu          sync.RWMutex
)

// translations holds all translation messages
var translations = map[Language]map[string]string{
	LangEN: messagesEN,
	LangZH: messagesZH,
}

// SetLanguage sets the current language
func SetLanguage(lang Language) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := translations[lang]; ok {
		currentLang = lang
	}
}

// GetLanguage returns the current language
func GetLanguage() Language {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// T returns the translated string for the given key
// Falls back to English if translation is not found
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if msgs, ok := translations[currentLang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Fallback to English
	if msgs, ok := translations[LangEN]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Return key if not found
	return key
}

// TWithLang returns the translated string for the given key in the specified language
func TWithLang(key string, lang Language) string {
	mu.RLock()
	defer mu.RUnlock()

	if msgs, ok := translations[lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Fallback to English
	if msgs, ok := translations[LangEN]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	return key
}

// SupportedLanguages returns a list of supported languages
func SupportedLanguages() []Language {
	return []Language{LangEN, LangZH}
}

// LanguageName returns the display name of a language
func LanguageName(lang Language) string {
	switch lang {
	case LangEN:
		return "English"
	case LangZH:
		return "中文"
	default:
		return string(lang)
	}
}
