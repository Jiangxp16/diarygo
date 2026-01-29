package i18n

import (
	"strings"
)

var lang = "en"
var mapTr map[string]string = nil

func Init(lang string) {
	lang = strings.ToLower(lang)
	var tmp map[string]string
	switch lang {
	case "zh", "zh-cn", "zhcn":
		tmp = Zh
	default:
		tmp = nil
	}
	if tmp != nil {
		mapTr = make(map[string]string, len(tmp))
		for k, v := range tmp {
			lowerKey := strings.ToLower(k)
			mapTr[lowerKey] = v
		}
	} else {
		mapTr = nil
	}
}

func T(key string) string {
	if mapTr == nil {
		return key
	}

	k := strings.ToLower(key)
	if v, ok := mapTr[k]; ok {
		return v
	}

	// 找不到翻译，回退 key
	return key
}
