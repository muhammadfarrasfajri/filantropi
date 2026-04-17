package model

type User struct {
	ID            string `json:"id"`
	IdToken       string `json:"id_token"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	WalletAddress string `json:"wallet_address"`
	Role          string `json:"role"`
	Isverified    int    `json:"is_verified"`
}
