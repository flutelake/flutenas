package trans

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
)

type Locale string

const (
	LocaleChinese    Locale = "zh-CN"
	LocaleEnglish    Locale = "en-US"
	LocaleRussian    Locale = "ru-RU"
	LocaleSpanish    Locale = "es-ES"
	LocaleFrench     Locale = "fr-FR"
	LocaleArabic     Locale = "ar-AR"
	LocalePortuguese Locale = "pt-PT"
)

// 当前环境中的语言
var defalutLocale = LocaleChinese

type Trans map[Locale]string

func (t Trans) Get(locale Locale) string {
	return t[locale]
}

// Default 自动获取环境中语言的翻译
func (t Trans) Default() string {
	return t[defalutLocale]
}

func (t Trans) Sprintf(locale Locale, args ...interface{}) string {
	return fmt.Sprintf(t[locale], args...)
}

var Langs = []Locale{LocaleChinese, LocaleEnglish, LocaleRussian, LocaleSpanish, LocaleFrench, LocaleArabic, LocalePortuguese}

// 保存所有翻译内容
var translation = map[string]Trans{}

// 中文翻译内容和ID的索引
var translationIndex = map[string]string{}

//go:embed i18n/*.txt
var f embed.FS

func init() {
	for _, lang := range Langs {
		data, err := f.Open(fmt.Sprintf("i18n/%s.txt", lang))
		if err != nil {
			panic(err)
		}

		transMap := map[string]string{}
		scanner := bufio.NewScanner(data)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
				continue
			}
			arr := strings.SplitN(line, ":", 2)
			if len(arr) != 2 {
				continue
			}
			if lang == LocaleChinese {
				translationIndex[arr[1]] = arr[0]
			}
			if v, ok := transMap[arr[0]]; ok {
				v = fmt.Sprintf("%s\n%s", v, strings.TrimSpace(arr[1]))
				transMap[arr[0]] = v
			} else {
				transMap[arr[0]] = strings.TrimSpace(arr[1])
			}
		}
		for id, content := range transMap {
			if ts, ok := translation[id]; ok {
				ts[lang] = content
				translation[id] = ts
			} else {
				translation[id] = Trans{lang: content}
			}
		}
	}
}

func SetDefaultLocale(locale string) {
	defalutLocale = Locale(locale)
}

func GetDefaultLocale() Locale {
	return defalutLocale
}

func GetTransMap(id string) Trans {
	return translation[id]
}

func GetTranslationMap(raw string) Trans {
	id, ok := translationIndex[raw]
	if !ok {
		return Trans{}
	}

	return GetTransMap(id)
}

// GetTranslation 通过 中文内容 获取其他语言对应的翻译内容， 注意： raw不支持多行内容
func GetTranslation(raw string, locale Locale) string {
	if locale == LocaleChinese {
		return raw
	}

	id, ok := translationIndex[raw]
	if !ok {
		return ""
	}

	content := GetTransMap(id).Get(locale)
	if content == "" {
		return raw
	}

	return content
}
