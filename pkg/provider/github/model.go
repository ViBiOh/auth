package github

type githubEmail struct {
	Email    string
	Primary  bool
	Verified bool
}

type githubUser struct {
	ID    uint
	Login string
}
