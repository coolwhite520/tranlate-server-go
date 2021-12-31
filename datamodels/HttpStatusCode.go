package datamodels

type HttpStatusCode int64

const (
	HttpSuccess HttpStatusCode =  200
	HttpDockerInitializing HttpStatusCode = -1000 - iota + 1
	HttpDockerRepairing
	HttpDockerServiceException
	HttpActivationNotFound    // 包括两个文件keystore、machineId文件
	HttpActivationReadFileError
	HttpActivationExpiredError
	HttpActivationGenerateError
	HttpActivationAESError
	HttpActivationInvalidateError
	HttpJwtTokenGenerateError
	HttpUserForbidden
	HttpUserNotLogin
	HttpUserTwicePwdNotSame
	HttpUserExpired
	HttpUserUpdatePwdError
	HttpUserPwdError
	HttpUserNameOrPwdError
	HttpFileTooBigger
	HttpUploadFileError
	HttpJsonParseError
	HttpLanguageNotSupport
	HttpTranslateError
	HttpRecordGetError
	HttpRecordDelError
	HttpUsersQueryError
	HttpUsersDeleteError
	HttpUsersAddError
	HttpUsersExistSameUserNameError
	HttpFileNotFoundError
	HttpFileOpenError
	// 新增的在后面添加
	HttpUserNoThisUserError
)

func (h HttpStatusCode) String() string {
	switch h {
	case HttpSuccess:
		return "成功"
	case HttpDockerInitializing:
		return "当前系统正在进行初始化,大约需要几分钟，请稍后"
	case HttpDockerRepairing:
		return  "当前系统服务异常，正在尝试自动修复，请稍后"
	case HttpDockerServiceException:
		return "当前系统服务异常，请联系管理员，进行系统修复。"
	case HttpActivationNotFound:
		return "未找到激活文件，请先激活后使用。"
	case HttpActivationReadFileError:
		return "激活文件读取错误"
	case HttpActivationExpiredError:
		return "激活码已过期"
	case HttpActivationGenerateError:
		return "生成激活文件失败"
	case HttpActivationAESError:
		return "激活文件AES错误"
	case HttpActivationInvalidateError:
		return "不是有效的激活码"
	case HttpUserForbidden:
		return "权限不足，禁止访问"
	case HttpUserNotLogin:
		return "用户未登录"
	case HttpUserTwicePwdNotSame:
		return "两次密码不一致，请重新输入"
	case HttpUserNoThisUserError:
		return "无此用户，请联系管理员进行开通"
	case HttpUserExpired:
		return "登录信息已过期"
	case HttpFileTooBigger:
		return "上传文件过大"
	case HttpJsonParseError:
		return "解析数据错误，参数传递错误"
	case HttpLanguageNotSupport:
		return "系统不支持传输的语言"
	default:
		return ""
	}
}
