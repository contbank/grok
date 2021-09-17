package grok

// PaginationResult ...
type PaginationResult struct {
	Total int64 `json:"total"`
	Pages int64 `json:"pages"`
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
