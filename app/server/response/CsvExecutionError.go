package response

type Error struct {
	LineNumber          int    `json:"lineNumber"`
	CustomerName        string `json:"customerName"`
	CustomerPhoneNumber string `json:"customerPhoneNumber"`
}

type CsvExectutionError struct {
	FileName string  `json:"fileName"`
	Errors   []Error `json:"errors"`
}
