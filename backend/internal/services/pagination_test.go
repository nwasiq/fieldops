package services

import "testing"

// rule: §3.2 — page_size is capped at 100 and the effective value is what the client sees
func TestNormalisePage(t *testing.T) {
	cases := []struct {
		page, size         int
		wantPage, wantSize int
	}{
		{0, 0, 1, DefaultPageSize},
		{-3, -1, 1, DefaultPageSize},
		{2, 50, 2, 50},
		{1, 100, 1, 100},
		{1, 101, 1, 100},
		{1, 100000, 1, 100},
	}
	for _, tc := range cases {
		got := NormalisePage(tc.page, tc.size)
		if got.Number != tc.wantPage || got.Size != tc.wantSize {
			t.Errorf("NormalisePage(%d, %d) = %+v, want page %d size %d", tc.page, tc.size, got, tc.wantPage, tc.wantSize)
		}
	}
	result := newPageResult([]int{1, 2, 3}, 250, NormalisePage(1, 100))
	if result.TotalPages != 3 {
		t.Errorf("250 rows at 100 per page = %d pages, want 3", result.TotalPages)
	}
	empty := newPageResult[int](nil, 0, NormalisePage(1, 20))
	if empty.TotalPages != 0 || empty.Items == nil {
		t.Errorf("empty result = %+v, want zero pages and a non-nil slice", empty)
	}
}
