package repeater

import (
	"errors"
	"fmt"
	"strings"
)

// lang 是用户可见消息的语言。
type lang string

const (
	langZH lang = "zh"
	langEN lang = "en"
)

func normalizeLang(raw string) lang {
	if strings.EqualFold(strings.TrimSpace(raw), "en") {
		return langEN
	}
	return langZH
}

func tr(l lang, zh, en string) string {
	if l == langEN {
		return en
	}
	return zh
}

func newLangError(l lang, zh, en string) error {
	return errors.New(tr(l, zh, en))
}

func newLangErrorf(l lang, zh, en string, args ...any) error {
	return fmt.Errorf(tr(l, zh, en), args...)
}
