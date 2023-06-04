package cpstructs

type Manifest struct {
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Version  string    `json:"version"`
	Owner    Owner     `json:"owner"`
	Archive  Archive   `json:"archive"`
	Licenses []License `json:"license"`
}

type Owner struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type Archive struct {
	Url  string `json:"url"`
	Md5  string `json:"md5"`
	Sha1 string `json:"sha1"`
}

type License struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}
