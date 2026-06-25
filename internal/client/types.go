package client

// Instance mirrors the nightlight-cloud utils.Instance struct.
type Instance struct {
	ID                   string                   `json:"id,omitempty"`
	Name                 string                   `json:"name"`
	InstanceType         string                   `json:"instanceType,omitempty"`
	Description          string                   `json:"description,omitempty"`
	InitializationStatus string                   `json:"initializationStatus,omitempty"`
	BootType             string                   `json:"bootType,omitempty"`
	CPUCores             int                      `json:"cpuCores"`
	CPUSockets           int                      `json:"cpuSockets"`
	MemoryMB             int                      `json:"memoryMB"`
	PrimaryIPAddress     string                   `json:"primaryIPAddress,omitempty"`
	PrimaryMacAddress    string                   `json:"primaryMacAddress,omitempty"`
	MetadataIPAddress    string                   `json:"metadataIPAddress,omitempty"`
	Devices              Devices                  `json:"devices"`
	PowerState           string                   `json:"powerState,omitempty"`
	ImageId              string                   `json:"imageId,omitempty"`
	SiteId               string                   `json:"siteId,omitempty"`
	DatastoreId          string                   `json:"datastoreId"`
	IPXEScript           string                   `json:"ipxeScript,omitempty"`
	SecureBoot           bool                     `json:"secureBoot"`
	TPM                  bool                     `json:"tpm"`
	StartPoweredOff      bool                     `json:"startPoweredOff"`
	UserData             string                   `json:"userData,omitempty"`
	VNCPort              int                      `json:"vncPort,omitempty"`
	Tags                 []map[string]interface{} `json:"tags"`
	CreatedAt            string                   `json:"createdAt,omitempty"`
	GuestStatus          string                   `json:"guestStatus,omitempty"`
	GuestAgentVersion    string                   `json:"guestAgentVersion,omitempty"`
	GuestHostname        string                   `json:"guestHostname,omitempty"`
	GuestOS              string                   `json:"guestOS,omitempty"`
	GuestIPAddresses     []string                 `json:"guestIPAddresses,omitempty"`
	GuestLastSeen        string                   `json:"guestLastSeen,omitempty"`
}

type Devices struct {
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces"`
	StorageDisks      []StorageDisk      `json:"storageDisks"`
	CDROMs            []CDROM            `json:"cdroms"`
	FloppyDisks       []FloppyDisk       `json:"floppyDisks"`
}

type NetworkInterface struct {
	ID          string `json:"id,omitempty"`
	IndexNumber int    `json:"indexNumber"`
	BootOrder   int    `json:"bootOrder"`
	Model       string `json:"model,omitempty"`
	Connected   bool   `json:"connected"`
	MacAddress  string `json:"macAddress,omitempty"`
	BridgeName  string `json:"bridgeName,omitempty"`
	SubnetId    string `json:"subnetId,omitempty"`
}

type StorageDisk struct {
	ID           string `json:"id,omitempty"`
	IndexNumber  int    `json:"indexNumber"`
	BootOrder    int    `json:"bootOrder"`
	SizeGB       int    `json:"sizeGB"`
	BusType      string `json:"busType,omitempty"`
	Path         string `json:"path,omitempty"`
	DatastoreId  string `json:"datastoreId,omitempty"`
	ExistingPath string `json:"existingPath,omitempty"`
	Clone        bool   `json:"clone"`
}

type CDROM struct {
	ID          string `json:"id,omitempty"`
	IndexNumber int    `json:"indexNumber"`
	BootOrder   int    `json:"bootOrder"`
	Connected   bool   `json:"connected"`
	Path        string `json:"path,omitempty"`
}

type FloppyDisk struct {
	IndexNumber int    `json:"indexNumber"`
	BootOrder   int    `json:"bootOrder"`
	Connected   bool   `json:"connected"`
	Path        string `json:"path,omitempty"`
}

// Subnet mirrors the nightlight-cloud Subnet struct.
type Subnet struct {
	ID          string                   `json:"id,omitempty"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	CIDRBlock   string                   `json:"cidrBlock"`
	SiteId      string                   `json:"siteId,omitempty"`
	VLANId      int64                    `json:"vlanId,omitempty"`
	DNSServers  []string                 `json:"dnsServers,omitempty"`
	NTPServers  []string                 `json:"ntpServers,omitempty"`
	DomainName  string                   `json:"domainName,omitempty"`
	BridgeName  string                   `json:"bridgeName,omitempty"`
	DHCPServer  bool                     `json:"dhcpServer"`
	IPPoolRange string                   `json:"ipPoolRange,omitempty"`
	Gateway     string                   `json:"gateway,omitempty"`
	Tags        []map[string]interface{} `json:"tags"`
}

// Datastore mirrors the nightlight-cloud Datastore struct.
type Datastore struct {
	ID          string                   `json:"id,omitempty"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Type        string                   `json:"type"`
	Path        string                   `json:"path,omitempty"`
	LocalPath   string                   `json:"localPath,omitempty"`
	Tags        []map[string]interface{} `json:"tags"`
}

// Site mirrors the nightlight-cloud Site struct.
type Site struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Location    string   `json:"location,omitempty"`
	Type        string   `json:"type,omitempty"`
	Topology    string   `json:"topology,omitempty"`
	Bridges     []string `json:"bridges,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Switch mirrors the nightlight-cloud Switch struct.
type Switch struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	SiteId      string   `json:"siteId,omitempty"`
	BridgeName  string   `json:"bridgeName,omitempty"`
	Type        string   `json:"type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Image mirrors the nightlight-cloud Image struct.
type Image struct {
	ID               string                   `json:"id,omitempty"`
	Name             string                   `json:"name"`
	Description      string                   `json:"description,omitempty"`
	Path             string                   `json:"path,omitempty"`
	Format           string                   `json:"format,omitempty"`
	SizeGB           float64                  `json:"sizeGB,omitempty"`
	OperatingSystem  string                   `json:"operatingSystem,omitempty"`
	Status           string                   `json:"status,omitempty"`
	DownloadProgress int                      `json:"downloadProgress,omitempty"`
	DatastoreId      string                   `json:"datastoreId,omitempty"`
	Tags             []map[string]interface{} `json:"tags"`
	CreatedAt        string                   `json:"createdAt,omitempty"`
}
