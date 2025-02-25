package schema

type Error struct {
	Message string `json:"message"`
	Details any    `json:"details"`
}

type Response[T any] struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Data   T      `json:"data"`
	Error  Error  `json:"error"`
}
