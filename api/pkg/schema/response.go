package schema

type Error struct {
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

type Response[T any] struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Data   T      `json:"data,omitempty"`
	Error  Error  `json:"error,omitempty"`
}
