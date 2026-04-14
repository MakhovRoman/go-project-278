package visits

type LinkVisit struct {
	ID        int64  `json:"id"`
	LinkID    int32  `json:"link_id"`
	Ip        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
	Status    int32  `json:"status"`
}
