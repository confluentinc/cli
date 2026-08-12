package switchover

// pairOut is the display structure for a switchover pair.
type pairOut struct {
	Id           string   `human:"ID" serialized:"id"`
	Name         string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment  string   `human:"Environment" serialized:"environment"`
	Members      []string `human:"Members" serialized:"members"`
	ActiveMember string   `human:"Active Member" serialized:"active_member"`
	Phase        string   `human:"Phase" serialized:"phase"`
}

// endpointOut is the display structure for a switchover endpoint.
type endpointOut struct {
	Id             string   `human:"ID" serialized:"id"`
	Name           string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment    string   `human:"Environment" serialized:"environment"`
	SwitchoverPair string   `human:"Switchover Pair" serialized:"switchover_pair"`
	Endpoints      []string `human:"Endpoints" serialized:"endpoints"`
	Target         string   `human:"Target,omitempty" serialized:"target,omitempty"`
	DrEndpoint     string   `human:"DR Endpoint,omitempty" serialized:"dr_endpoint,omitempty"`
	Phase          string   `human:"Phase" serialized:"phase"`
}
