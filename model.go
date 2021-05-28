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

	BatchUploadFolder struct {
		Name   string
		Number int
		Path   string
	}

	ListPostResponse struct {
		Query   string `json:"query"`
		Offset  int    `json:"offset"`
		Limit   int    `json:"limit"`
		Total   int    `json:"total"`
		Results []Post `json:"results"`
	}
)
