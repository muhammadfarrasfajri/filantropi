package model

type APIResponse struct {
	Error   bool        `json:"error"`          // True/False
	Message string      `json:"message"`        // Pesan untuk user
	Type    string      `json:"type,omitempty"` // Jenis error (ValidationError, etc)
	Data    interface{} `json:"data,omitempty"` // Data sukses (opsional)
}
