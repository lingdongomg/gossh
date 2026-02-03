package i18n

import (
	"testing"
)

func TestSetAndGetLanguage(t *testing.T) {
	// Save original
	original := GetLanguage()
	defer SetLanguage(original)

	tests := []struct {
		name string
		lang Language
	}{
		{"English", LangEN},
		{"Chinese", LangZH},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLanguage(tt.lang)
			got := GetLanguage()
			if got != tt.lang {
				t.Errorf("GetLanguage() = %v, want %v", got, tt.lang)
			}
		})
	}
}

func TestSetLanguageInvalid(t *testing.T) {
	// Save original
	original := GetLanguage()
	defer SetLanguage(original)

	SetLanguage(LangEN)
	SetLanguage(Language("invalid"))
	
	// Should still be EN (invalid ignored)
	if GetLanguage() != LangEN {
		t.Error("Invalid language should be ignored")
	}
}

func TestT(t *testing.T) {
	// Save original
	original := GetLanguage()
	defer SetLanguage(original)

	tests := []struct {
		name string
		lang Language
		key  string
		want string
	}{
		{"English welcome", LangEN, "app.welcome", "Welcome to GoSSH"},
		{"Chinese welcome", LangZH, "app.welcome", "欢迎使用 GoSSH"},
		{"English quit", LangEN, "menu.quit", "Quit"},
		{"Chinese quit", LangZH, "menu.quit", "退出"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLanguage(tt.lang)
			got := T(tt.key)
			if got != tt.want {
				t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestTFallback(t *testing.T) {
	// Save original
	original := GetLanguage()
	defer SetLanguage(original)

	// Test that unknown key returns the key itself
	SetLanguage(LangEN)
	key := "unknown.key.that.does.not.exist"
	got := T(key)
	if got != key {
		t.Errorf("T(%q) = %q, want %q (key itself)", key, got, key)
	}
}

func TestTWithLang(t *testing.T) {
	// Should work regardless of current language
	en := TWithLang("app.welcome", LangEN)
	zh := TWithLang("app.welcome", LangZH)

	if en != "Welcome to GoSSH" {
		t.Errorf("TWithLang(EN) = %q, want %q", en, "Welcome to GoSSH")
	}

	if zh != "欢迎使用 GoSSH" {
		t.Errorf("TWithLang(ZH) = %q, want %q", zh, "欢迎使用 GoSSH")
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := SupportedLanguages()

	if len(langs) != 2 {
		t.Errorf("SupportedLanguages() returned %d languages, want 2", len(langs))
	}

	hasEN := false
	hasZH := false
	for _, lang := range langs {
		if lang == LangEN {
			hasEN = true
		}
		if lang == LangZH {
			hasZH = true
		}
	}

	if !hasEN {
		t.Error("SupportedLanguages should include English")
	}
	if !hasZH {
		t.Error("SupportedLanguages should include Chinese")
	}
}

func TestLanguageName(t *testing.T) {
	tests := []struct {
		lang Language
		want string
	}{
		{LangEN, "English"},
		{LangZH, "中文"},
		{Language("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			got := LanguageName(tt.lang)
			if got != tt.want {
				t.Errorf("LanguageName(%v) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}

func TestAllKeysExistInBothLanguages(t *testing.T) {
	// Get all keys from English
	for key := range messagesEN {
		if _, ok := messagesZH[key]; !ok {
			t.Errorf("Key %q exists in English but not in Chinese", key)
		}
	}

	// Get all keys from Chinese
	for key := range messagesZH {
		if _, ok := messagesEN[key]; !ok {
			t.Errorf("Key %q exists in Chinese but not in English", key)
		}
	}
}
