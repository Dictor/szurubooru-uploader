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
		Version      int           `json:"version"`
		Id           int           `json:"id"`
		Safety       string        `json:"safety"`
		Tags         []interface{} `json:"tags"`
		ThumbnailUrl string        `json:"thumbnailUrl"`
	}

	BatchUploadFolder struct {
		Name   string
		Number int
		Path   string
	}
)
