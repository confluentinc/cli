package shared

type User struct {
	ID        int      `json:"id" hcl:"id"`
	Email     string   `json:"email" hcl:"email"`
	FirstName string   `json:"first_name" hcl:"first_name"`
	LastName  string   `json:"last_name" hcl:"last_name"`
	OrgID     int      `json:"organization_id" hcl:"organization_id"`
}

type Account struct {
	ID        string   `json:"id" hcl:"id"`
	Name      string   `json:"name" hcl:"name"`
	OrgID     int      `json:"organization_id" hcl:"organization_id"`
}

// API: GET /me
type AuthConfig struct {
	User      *User    `json:"user" hcl:"user"`
	Account   *Account `json:"account" hcl:"account"`
}
