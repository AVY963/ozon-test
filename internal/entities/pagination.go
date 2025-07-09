package entities

type PaginationRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginationResponse struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

func NewPaginationRequest(limit, offset int) *PaginationRequest {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return &PaginationRequest{
		Limit:  limit,
		Offset: offset,
	}
}

func NewPaginationResponse(total int64, limit, offset int) *PaginationResponse {
	return &PaginationResponse{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+limit) < total,
	}
}
