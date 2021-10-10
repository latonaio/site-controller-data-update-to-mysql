package response

type AuthCSV struct {
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Path      string `json:"path"`
}
