package handler

import "strconv"

// parseID преобразует строковый ID из URL в int64.
// Используется во всех handler'ах, где нужен числовой ID из chi.URLParam.
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
