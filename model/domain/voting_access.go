package domain

type VotingAccess struct {
	UserId int    `json:"user_id"`
	Hashed string `json:"hashed"`
}
