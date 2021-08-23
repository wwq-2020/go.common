package rpc

// Wraper Wraper
type Wraper interface {
	Wrap() bool
}

// UnWraper UnWraper
type UnWraper interface {
	UnWrap() bool
}

func isRespNeedWrap(resp interface{}) bool {
	if resp == nil {
		return false
	}
	wrapper, isWraper := resp.(Wraper)
	return isWraper && wrapper.Wrap()
}

func isRespNeedUnWrap(resp interface{}) bool {
	if resp == nil {
		return false
	}
	unWrapper, isUnWraper := resp.(UnWraper)
	return isUnWraper && unWrapper.UnWrap()
}
