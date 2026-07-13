package client

// Command is the payload for the sendkeys console API.
type Command struct {
	KeyCode    string `json:"keyCode"`
	RawMapping bool   `json:"rawMapping,omitempty"`
	RawKeyCode uint32 `json:"rawKeyCode,omitempty"`
}

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
	WinAutoattend        string                   `json:"winAutoattend,omitempty"`
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

// RouterConfig mirrors the nightlight-cloud utils.RouterConfig struct.
type RouterConfig struct {
	InstanceID   string            `json:"instanceId,omitempty"`
	RouterType   string            `json:"routerType,omitempty"`
	Hostname     string            `json:"hostname,omitempty"`
	Interfaces   []RouterInterface `json:"interfaces"`
	StaticRoutes []StaticRoute     `json:"staticRoutes"`
	Subnets      []RouterSubnet    `json:"subnets,omitempty"`
	NAT          NATConfig         `json:"nat"`
}

// RouterInterface mirrors the nightlight-cloud utils.RouterInterface struct.
type RouterInterface struct {
	MacAddress string `json:"macAddress,omitempty"`
	IPAddress  string `json:"ipAddress,omitempty"`
	Netmask    string `json:"netmask,omitempty"`
	Gateway    string `json:"gateway,omitempty"`
}

// RouterSubnet mirrors the nightlight-cloud utils.RouterSubnet struct.
type RouterSubnet struct {
	SubnetID  string `json:"subnetId,omitempty"`
	Name      string `json:"name,omitempty"`
	CIDRBlock string `json:"cidrBlock,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
	VLANId    int64  `json:"vlanId,omitempty"`
}

// NATConfig mirrors the nightlight-cloud utils.NATConfig struct.
type NATConfig struct {
	Enabled   bool   `json:"enabled"`
	Interface string `json:"interface,omitempty"`
}

// WanRouterConfig mirrors the nightlight-cloud WanRouterConfig struct.
type WanRouterConfig struct {
	ID           string        `json:"id,omitempty"`
	WANIPAddress string        `json:"wanIPAddress"`
	WANNetmask   string        `json:"wanNetmask"`
	WANGateway   string        `json:"wanGateway"`
	StaticRoutes []StaticRoute `json:"staticRoutes"`
	NATEnabled   bool          `json:"natEnabled"`
}

// StaticRoute mirrors the nightlight-cloud utils.StaticRoute struct.
type StaticRoute struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
}

// ContentLibrary mirrors the nightlight-cloud ContentLibrary struct.
type ContentLibrary struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	DepotURL     string `json:"depotUrl"`
	DepotToken   string `json:"depotToken,omitempty"`
	DatastoreId  string `json:"datastoreId"`
	SyncInterval string `json:"syncInterval,omitempty"`
	SyncStatus   string `json:"syncStatus,omitempty"`
	SyncError    string `json:"syncError,omitempty"`
	LastSyncAt   string `json:"lastSyncAt,omitempty"`
	ItemCount    int    `json:"itemCount,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// UpdateContentLibraryRequest mirrors the nightlight-cloud updateContentLibraryRequest struct.
type UpdateContentLibraryRequest struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description"`
	DepotURL     string `json:"depotUrl,omitempty"`
	DepotToken   string `json:"depotToken,omitempty"`
	SyncInterval string `json:"syncInterval,omitempty"`
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
