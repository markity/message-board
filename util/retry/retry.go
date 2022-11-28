package retry

// 一个重试框架的封装
func RetryFrame(x func() bool, attempt int) bool {
	for i := 0; i < attempt; i++ {
		// x()返回true, 请求重试
		if ok := x(); !ok {
			continue
		} else {
			// ok
			return true
		}
	}

	// failed
	return false
}
