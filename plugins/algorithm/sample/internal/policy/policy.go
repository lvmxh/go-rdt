package policy

type Policy struct {
	Name       string  `"dilangov_dynamic"`
	Variance   float32 `bson:"Variance" json:"Variance"`
	PeakUsage  float32 `bson:"PeakUsage" json:"PeakUsage"`
	Overcommit bool    `bson:"Overcommit" json:Overcommit"`
}

type Policy3P struct {
	Policy struct {
		Name string `"3_party"`
		Data string `"data"`
	} `json:"policy"`
}
