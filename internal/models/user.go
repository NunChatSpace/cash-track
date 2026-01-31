package models

type User struct {
	ID        int64 `json:"id"`
	Name      string `json:"name"`
	CutoffDay int    `json:"cutoff_day"`
}
