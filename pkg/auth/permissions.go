package auth

var (
	// Device permissions
	READ_DEVICES  permission = "READ_DEVICES"
	WRITE_DEVICES permission = "WRITE_DEVICES"

	// API Key permissions
	READ_API_KEYS  permission = "READ_API_KEYS"
	WRITE_API_KEYS permission = "WRITE_API_KEYS"
)

type permission string

func (p permission) String() string {
	return string(p)
}
