package pagination

import "gorm.io/gorm"

// PaginateResult represents the pagination metadata and data.
//
// @Description This struct contains metadata about the pagination process
// along with the data retrieved for the current page.
type PaginateResult struct {
	Data        interface{} `json:"data"`         // The data for the current page
	CurrentPage int         `json:"current_page"` // The current page number
	From        int         `json:"from"`         // The starting record number for the current page
	To          int         `json:"to"`           // The ending record number for the current page
	LastPage    int         `json:"last_page"`    // The total number of pages
	PerPage     int         `json:"per_page"`     // The number of records per page
	Total       int64       `json:"total"`        // The total number of records
}

// Paginate performs pagination on a GORM query.
//
// @Description This function applies pagination logic to a GORM database query. 
// It allows optional custom query modifications via the `rawFunc` parameter and 
// retrieves a subset of data based on the specified page and limit.
//
// @Tags Pagination
//
// @param db *gorm.DB - The GORM database instance
// @param page int - The current page number (1-based index)
// @param limit int - The maximum number of records per page
// @param rawFunc func(*gorm.DB) *gorm.DB - Optional custom query modifier function
// @param output interface{} - A pointer to the slice where the query results will be stored
//
// @return PaginateResult - A struct containing paginated data and metadata
// @return error - An error if the query fails
func Paginate(db *gorm.DB, page, limit int, rawFunc func(*gorm.DB) *gorm.DB, output interface{}) (PaginateResult, error) {
	// Calculate the offset for the current page
	offset := (page - 1) * limit

	// Start with the base query
	query := db
	if rawFunc != nil {
		// Apply optional custom query modifications
		query = rawFunc(query)
	}

	// Count the total number of records
	var total int64
	query.Model(output).Count(&total)

	// Execute the query with offset and limit
	err := query.Offset(offset).Limit(limit).Find(output).Error
	if err != nil {
		// Return an empty result if the query fails
		return PaginateResult{}, err
	}

	// Calculate the ending record number for the current page
	to := offset + limit
	if to > int(total) {
		to = int(total) // Adjust if the total records are less than the limit
	}

	// Return the paginated result
	return PaginateResult{
		Data:        output,                  // The data for the current page
		CurrentPage: page,                    // The current page number
		From:        offset + 1,              // The starting record number
		To:          to,                      // The ending record number
		LastPage:    (int(total) + limit - 1) / limit, // Total pages (ceil(total/limit))
		PerPage:     limit,                   // Records per page
		Total:       total,                   // Total records
	}, nil
}
