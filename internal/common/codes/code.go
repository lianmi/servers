package codes

//错误码定义
const (
	SUCCESS       = 200
	ERROR         = 500
	InvalidParams = 400
	NONEREGISTER  = 404

	ErrExistTag        = 10001
	ErrNotExistTag     = 10002
	ErrNotExistArticle = 10003

	ErrAuthCheckTokenFail    = 20001
	ErrAuthCheckTokenTimeout = 20002
	ErrAuthToken             = 20003
	ErrAuth                  = 20004

	ErrNotDigital   = 29001
	ErrWrongSmsCode = 29002

	ErrExistUser         = 30001
	ErrExistMobile       = 30002
	ErrNotRegisterMobile = 30003
	ErrNotFoundInviter   = 30004

	PageNotFound = 40001
)
