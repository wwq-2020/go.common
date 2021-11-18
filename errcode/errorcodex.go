package errcode

func CodeIs(code int32, errcode ErrCode) bool {
	return ErrCode(code) == errcode
}

func (e ErrCode) Code() int32 {
	return int32(e)
}
