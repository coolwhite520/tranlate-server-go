package structs

import (
	"github.com/mozillazg/go-pinyin"
)

type SupportLang struct {
	EnName string `json:"en_name"`
	CnName string `json:"cn_name"`
}

func Hans2Pinyin(hans string) string {
	args := pinyin.NewArgs()
	rows := pinyin.Pinyin(hans, args)

	strResult := ""
	for _, v := range rows {
		for _, item := range v {
			strResult += item
		}
	}
	return strResult
}

type LangList []SupportLang

func (a LangList) Len() int      { return len(a) }
func (a LangList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a LangList) Less(i, j int) bool {
	return Hans2Pinyin(a[i].CnName) < Hans2Pinyin(a[j].CnName)
}

type Activation struct {
	UserName        string   `json:"user_name"`
	Sn              string   `json:"sn"`
	SupportLangList LangList `json:"support_lang_list"` // 英文简称列表
	CreatedAt       int64    `json:"created_at"`        // 这个时间代表激活码的生成时间，授权人员生成激活码的时候决定了
	UseTimeSpan     int64    `json:"use_time_span"`     // 可以使用的时间，是一个时间段，以秒为单位 比如一年：1 * 365 * 24 * 60 * 60
	Mark            string   `json:"mark"`
}

// KeystoreLeftTime
// 验证流程：
// 1。不存在安装路径下的keystore和/usr/bin/${machineID} 两个文件
// 2。第一次激活后生成keystore、/usr/bin/${machineID}文件，文件内容为KeystoreLeftTime结构体
// 3。启动线程，每隔10分钟读取/usr/bin/${machineID}文件，然后减少里面的LeftTimeSpan - 10（分钟）并重新写回去
// 4。在activation_middleware.go中进行解析，并判断LeftTimeSpan是否大于0，如果小于0，提示用户过期（考虑到多线程对文件的读写，所以需要使用channel进行mutex的控制）
// 特殊情况：
// 1。 如果已经存在/usr/bin/${machineID}文件，keystore丢失的时候，
//     那么只有当新的激活码中的CreatedAt和已经存在/usr/bin/${machineID}文件中的
//     CreatedAt不同的时候才会重新生成/usr/bin/${machineID}文件，否则继续使用
// 2。 如果存在keystore，而不存在/usr/bin/${machineID}的时候，这种属于用户删除的行为。没有办法只能重新生成
//    /usr/bin/${machineID}文件，然后 LeftTimeSpan= UseTimeSpan - （当前时间 - CreatedAt）

type KeystoreExpired struct {
	Sn           string `json:"sn"`
	CreatedAt    int64  `json:"created_at"`
	LeftTimeSpan int64  `json:"left_time_span"` // 初始化为UseTimeSpan
}

// BannedKeystoreInfo 失效的授权id(CreatedAt)列表
type BannedKeystoreInfo struct {
	Ids []int64 `json:"ids"`
}

type ProofStatusCode int

const (
	ProofStateOk ProofStatusCode = iota
	ProofStateNotActivation
	ProofStateExpired
	ProofStateForceBanned
)

func (p ProofStatusCode) String() string {
	switch p {
	case ProofStateOk:
		return "使用中"
	case ProofStateForceBanned:
		return "强制失效"
	case ProofStateExpired:
		return "已经过期"
	case ProofStateNotActivation:
		return "首次激活"
	default:
		return ""
	}
}

// KeystoreProof 凭证
type KeystoreProof struct {
	Sn            string          `json:"sn"`    // 机器码
	State         ProofStatusCode `json:"state"` // 授权状态 0。使用中 1。未激活 2。过期 3。强制失效
	StateDescribe string          `json:"state_describe"`
	Now           int64           `json:"now"` // 时间以便每次生成都不同
}
