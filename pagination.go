package grok

// PaginationResult ...
type PaginationResult struct {
	Total         int64 `json:"total"`
	TotalReturned int64 `json:"total_returned"`
	PerPage       int64 `json:"per_page"`
	CurrentPage   int64 `json:"current_page"`
	Pages         int64 `json:"pages"`
	HasNextPage   bool  `json:"has_next_page"`
}

// TotalPage ...
// totalRows: recebe total registros ex.: 687
// perPage: recebe total de registros por pagina ex.: 100
func TotalPage(totalRows int64, perPage int64) (pag int64) {
	totPage := totalRows / perPage

	rest := totalRows % perPage
	if rest > 0 {
		totPage += 1
	}

	return totPage
}
