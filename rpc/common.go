package rpc

// Wraper Wraper
type Wraper interface {
	Wrap() bool
}

func isRespNeedWrap(resp interface{}) bool {
	if resp == nil {
		return false
	}
	wrapper, isWraper := resp.(Wraper)
	return isWraper && wrapper.Wrap()
}
