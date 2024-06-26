package main

type (
	ReverseSearchResponse struct {
		ExactPost    Post          `json:"exactPost`
		SimilarPosts []SimilarPost `json:"similarPosts`
	}

	SimilarPost struct {
		Distance float64 `json:"distance"`
		Post     Post    `json:"post"`
	}

	Post struct {
		Version       int           `json:"version"`
		Id            int           `json:"id"`
		Safety        string        `json:"safety"`
		Tags          []interface{} `json:"tags"`
		ThumbnailUrl  string        `json:"thumbnailUrl"`
		FavoriteCount int           `json:"favoriteCount"`
	}

	Tag struct {
		Version      int      `json:"version"`
		Names        []string `json:"names"`
		Category     string   `json:"category"`
		Implications []Tag    `json:"implications"`
		Suggestions  []Tag    `json:"suggestions"`
		Usages       int      `json:"usages"`
		Description  string   `json:"description"`
	}

	ImplicationUpdateRequest struct {
		Version      int      `json:"version"`
		Implications []string `json:"implications"`
	}

	BatchUploadFolder struct {
		Name   string
		Number int
		Path   string
	}

	ListResponse struct {
		Query  string `json:"query"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
		Total  int    `json:"total"`
	}

	ListPostResponse struct {
		ListResponse
		Results []Post `json:"results"`
	}

	ListTagResponse struct {
		ListResponse
		Results []Tag `json:"results"`
	}
)
