package dumper

type Type string

const (
	TypePostgres       Type = "postgres"
	TypeMysql          Type = "mysql"
	TypeMongo          Type = "mongo"
	TypeMongoLegacy    Type = "mongo_legacy"
	TypeFirebirdLegacy Type = "firebird_legacy"
	TypeTar            Type = "tar"
)

type GlobalConfiguration struct {
	Path                 string `yaml:"path"`
	TmpPath              string `yaml:"tmp-path"`
	ShExecutable         string `yaml:"sh-executable"`
	PgdumpExecutable     string `yaml:"pgdump-executable"`
	MysqldumpExecutable  string `yaml:"mysqldump-executable"`
	Mongodump5Executable string `yaml:"mongodump-5-executable"`
	Mongodump4Executable string `yaml:"mongodump-4-executable"`
	GbakExecutable       string `yaml:"gbak-executable"`
	TarExecutable        string `yaml:"tar-executable"`
}

type Configuration struct {
	Type Type   `yaml:"type"`
	Name string `yaml:"name"`

	//override destination path
	//when empty, path will be global path + dumper name
	Path string `yaml:"path"`

	//override tmp path
	TmpPath string `yaml:"tmp-path"`

	//variables to pass to dump executable
	Vars map[string]string `yaml:"vars"`

	//always make the latest dump, even if daily/weekly/monthly dumps exist
	ForceLatest bool `yaml:"force-latest"`

	//make daily dumps
	Daily bool `yaml:"daily"`
	Days  int  `yaml:"days"`

	//make weekly dumps
	Weekly bool `yaml:"weekly"`
	Weeks  int  `yaml:"weeks"`

	//make monthly dumps
	Monthly bool `yaml:"monthly"`
	Months  int  `yaml:"months"`
}
