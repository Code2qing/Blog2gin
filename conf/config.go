package conf

type Database struct {
	Type        string `json:"type"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Name        string `json:"name"`
	TablePrefix string `json:"table_prefix"`
}

type Config struct {
	Address  string   `json:"address"`
	Port     int      `json:"port"`
	Database Database `json:"database"`
}
