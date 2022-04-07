package constant

import (
	"github.com/Unknwon/goconfig"
	log "github.com/sirupsen/logrus"
	"sort"
	"translate-server/structs"
)

const contentAll = `
[lang]
Afrikaans=南非荷兰语
Amharic=阿姆哈拉语
Arabic=阿拉伯语
Asturian=阿斯图里亚斯语
Azerbaijani=阿塞拜疆语
Bashkir=巴什基尔语
Belarusian=白俄罗斯语
Bulgarian=保加利亚语
Bengali=孟加拉语
Breton=布列塔尼语
Bosnian=波斯尼亚语
Catalan=加泰罗尼亚语
Valencian=瓦伦西亚语
Cebuano=宿雾语
Czech=捷克语
Welsh=威尔士语
Danish=丹麦语
German=德语
Greek=希腊语
English=英语
Spanish=西班牙语
Estonian=爱沙尼亚语
Persian=波斯语
Fulah=福拉语
Finnish=芬兰语
French=法语
"Western Frisian"=西弗里斯兰语
Irish=爱尔兰语
Gaelic=盖尔语
"Scottish Gaelic"=苏格兰盖尔语
Galician=加利西亚语
Gujarati=古吉拉特语
Hausa=豪萨语
Hebrew=希伯来语
Hindi=印地语
Croatian=克罗地亚语
Haitian=海地语
"Haitian Creole"=海地克里奥尔语
Hungarian=匈牙利语
Armenian=亚美尼亚语
Indonesian=印度尼西亚语
Igbo=伊博语
Iloko=伊洛科语
Icelandic=冰岛语
Italian=意大利语
Japanese=日语
Javanese=爪哇语
Georgian=格鲁吉亚语
Kazakh=哈萨克语
Khmer=高棉语
"Central Khmer"=中央高棉语
Kannada=卡纳达语
Korean=韩语
Luxembourgish=卢森堡语
Letzeburgesch=莱茨堡语
Ganda=甘达语
Lingala=林加拉语
Lao=老挝语
Lithuanian=立陶宛语
Latvian=拉脱维亚语
Malagasy=马尔加什语
Macedonian=马其顿语
Malayalam=马拉雅拉姆语
Mongolian=蒙语
Marathi=马拉地语
Malay=马来语
Burmese=缅甸语
Nepali=尼泊尔语
Dutch=荷兰语
Flemish=佛兰芒语
Norwegian=挪威语
"Northern Sotho"=北索托语
Occitan=奥克西坦语
Oriya=奥里亚语
Panjabi=旁遮普语
;Punjabi=旁遮普语
Polish=抛光语
Pushto=普什图语
;Pashto=普什图语
Portuguese=葡萄牙语
Romanian=罗马尼亚语
Moldavian=摩尔多瓦语
;Moldovan=摩尔多瓦语
Russian=俄语
Sindhi=信德语
;Sinhala=僧伽罗语
Sinhalese=僧伽罗语
Slovak=斯洛伐克语
Slovenian=斯洛文尼亚语
Somali=索马里语
Albanian=阿尔巴尼亚语
Serbian=塞尔维亚语
Swati=斯瓦蒂语
Sundanese=巽他语
Swedish=瑞典语
Swahili=斯瓦希里语
Tamil=泰米尔语
Thai=泰国语
Tagalog=他加禄语
Tswana=茨瓦纳语
Turkish=土耳其语
Ukrainian=乌克兰语
Urdu=乌尔都语
Uzbek=乌兹别克语
Vietnamese=越南语
Wolof=沃洛夫语
Xhosa=科萨语
Yiddish=意第绪语语
Yoruba=约鲁巴语
Chinese=中文(简体)
Zulu=祖鲁语
`

var AllLanguageList structs.LangList

func init() {
	cfg, err := goconfig.LoadFromData([]byte(contentAll))
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	section, err := cfg.GetSection("lang")
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	for k, v := range section{
		AllLanguageList = append(AllLanguageList, structs.SupportLang{
			EnName: k,
			CnName: v,
		})
	}
	sort.Sort(AllLanguageList)
}