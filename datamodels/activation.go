package datamodels

type SupportLang struct {
	EnName string `json:"en_name"`
	CnName string `json:"cn_name"`
}

type Activation struct {
	UserName        string        `json:"user_name"`
	Sn              string        `json:"sn"`
	SupportLangList []SupportLang `json:"support_lang_list"` // 英文简称列表
	CreatedAt       int64         `json:"created_at"`        // 这个时间代表激活码的生成时间，授权人员生成激活码的时候决定了
	UseTimeSpan     int64         `json:"use_time_span"`     // 可以使用的时间，是一个时间段，以秒为单位 比如一年：1 * 365 * 24 * 60 * 60
	Mark            string        `json:"mark"`
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
	Sn           string  `json:"sn"`
	CreatedAt    int64   `json:"created_at"`
	LeftTimeSpan int64  `json:"left_time_span"`   // 初始化为UseTimeSpan
}
