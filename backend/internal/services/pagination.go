package services

// Pagination bounds (conventions §8).
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Page is a normalised page request.
type Page struct {
	Number int
	Size   int
}

// NormalisePage clamps a raw page request: page defaults to 1, page_size to
// DefaultPageSize, and page_size never exceeds MaxPageSize. The returned
// values are the effective ones and must be echoed to the client.
func NormalisePage(number, size int) Page {
	if number < 1 {
		number = 1
	}
	switch {
	case size < 1:
		size = DefaultPageSize
	case size > MaxPageSize:
		size = MaxPageSize
	}
	return Page{Number: number, Size: size}
}

// Offset is the row offset for the page.
func (p Page) Offset() int {
	return (p.Number - 1) * p.Size
}

// PageResult is one page of a list plus the numbers a client needs to page on.
type PageResult[T any] struct {
	Items      []T
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

func newPageResult[T any](items []T, total int64, page Page) PageResult[T] {
	totalPages := int((total + int64(page.Size) - 1) / int64(page.Size))
	if items == nil {
		items = []T{}
	}
	return PageResult[T]{Items: items, Total: total, Page: page.Number, PageSize: page.Size, TotalPages: totalPages}
}
