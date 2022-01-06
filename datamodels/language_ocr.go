package datamodels

import (
	"github.com/Unknwon/goconfig"
	log "github.com/sirupsen/logrus"
)

const content = `
[language_ocr]
"Central Khmer"=khm
"Haitian Creole"=hat
"Northern Sotho"=afr
"Scottish Gaelic"=ind
"Western Frisian"=deu
Afrikaans=afr
Albanian=sqi
Amharic=amh
Arabic=ara
Armenian=afr
Asturian=spa
Azerbaijani=aze
Bashkir=rus
Belarusian=bel
Bengali=ben
Bosnian=bos
Breton=bre
Bulgarian=bul
Burmese=mya
Catalan=cat
Cebuano=ceb
Chinese=chi_sim
Croatian=hrv
Czech=ces
Danish=dan
Dutch=nld
English=eng
Estonian=est
Finnish=fin
Flemish=nld
French=fra
Fulah=afr
Gaelic=gle
Galician=glg
Ganda=ind
Georgian=kat
German=deu
Greek=ell
Gujarati=guj
Haitian=hat
Hausa=lat
Hebrew=heb
Hindi=hin
Hungarian=hun
Icelandic=isl
Igbo=afr
Iloko=ind
Indonesian=ind
Irish=eng
Italian=ita
Japanese=jpn
Javanese=jav
Kannada=kan
Kazakh=kaz
Khmer=khm
Korean=kor
Lao=lao
Latvian=lat
Letzeburgesch=deu
Lingala=afr
Lithuanian=lit
Luxembourgish=deu
Macedonian=mkd
Malagasy=msa
Malay=msa
Malayalam=mal
Marathi=mar
Moldavian=ron
Moldovan=ron
Mongolian=mon
Nepali=nep
Norwegian=nor
Occitan=oci
Oriya=ori
Panjabi=pan
Pashto=pus
Persian=fas
Polish=pol
Portuguese=por
Punjabi=pan
Pushto=pus
Romanian=ron
Russian=rus
Serbian=srp
Sindhi=hin
Sinhala=sin
Sinhalese=sin
Slovak=slk
Slovenian=slv
Somali=ell
Spanish=spa
Sundanese=ind
Swahili=swa
Swati=swa
Swedish=swe
Tagalog=tgl
Tamil=tam
Thai=tha
Tswana=por
Turkish=tur
Ukrainian=ukr
Urdu=urd
Uzbek=uzb
Valencian=cat
Vietnamese=vie
Welsh=cym
Wolof=msa
Xhosa=afr
Yiddish=yid
Yoruba=afr
Zulu=afr
`

var LanguageOcrList map[string]string

func init() {
	cfg, err := goconfig.LoadFromData([]byte(content))
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	section, err := cfg.GetSection("language_ocr")
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	LanguageOcrList = section
}