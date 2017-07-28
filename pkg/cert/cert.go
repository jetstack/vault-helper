package cert

type Cert struct {
	role        string
	commonName  string
	destination string
	bitSize     int
	key         string
	ipSans      []string
	sanHosts    []string
	owner       string
	group       string
}
