package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaginationRequest(t *testing.T) {
	testCases := []struct {
		name           string
		inputLimit     int
		inputOffset    int
		expectedLimit  int
		expectedOffset int
	}{
		{"valid_values", 10, 5, 10, 5},
		{"zero_limit", 0, 5, 20, 5},
		{"negative_limit", -5, 5, 20, 5},
		{"too_high_limit", 150, 5, 20, 5},
		{"negative_offset", 10, -5, 10, 0},
		{"zero_offset", 10, 0, 10, 0},
		{"both_invalid", -10, -5, 20, 0},
		{"max_valid_limit", 100, 50, 100, 50},
		{"just_over_limit", 101, 0, 20, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := NewPaginationRequest(tc.inputLimit, tc.inputOffset)

			assert.Equal(t, tc.expectedLimit, req.Limit)
			assert.Equal(t, tc.expectedOffset, req.Offset)
		})
	}
}

func TestNewPaginationResponse(t *testing.T) {
	testCases := []struct {
		name            string
		total           int64
		limit           int
		offset          int
		expectedHasMore bool
	}{
		{"no_more", 10, 20, 0, false},
		{"has_more", 100, 20, 0, true},
		{"exact_end", 20, 20, 0, false},
		{"middle_page", 100, 20, 20, true},
		{"last_page", 100, 20, 80, false},
		{"empty_result", 0, 20, 0, false},
		{"single_item", 1, 20, 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := NewPaginationResponse(tc.total, tc.limit, tc.offset)

			assert.Equal(t, tc.total, resp.Total)
			assert.Equal(t, tc.limit, resp.Limit)
			assert.Equal(t, tc.offset, resp.Offset)
			assert.Equal(t, tc.expectedHasMore, resp.HasMore)
		})
	}
}

func TestPaginationResponseHasMore(t *testing.T) {
	resp1 := NewPaginationResponse(100, 20, 0)
	assert.True(t, resp1.HasMore)

	resp2 := NewPaginationResponse(100, 20, 80)
	assert.False(t, resp2.HasMore)

	resp3 := NewPaginationResponse(20, 20, 0)
	assert.False(t, resp3.HasMore)

	resp4 := NewPaginationResponse(0, 20, 0)
	assert.False(t, resp4.HasMore)
}

func TestPaginationRequestFields(t *testing.T) {
	req := &PaginationRequest{
		Limit:  15,
		Offset: 30,
	}

	assert.Equal(t, 15, req.Limit)
	assert.Equal(t, 30, req.Offset)
}

func TestPaginationResponseFields(t *testing.T) {
	resp := &PaginationResponse{
		Total:   150,
		Limit:   25,
		Offset:  50,
		HasMore: true,
	}

	assert.Equal(t, int64(150), resp.Total)
	assert.Equal(t, 25, resp.Limit)
	assert.Equal(t, 50, resp.Offset)
	assert.True(t, resp.HasMore)
}
