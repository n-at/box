package dumper

type Type string

const (
	TypePostgres       Type = "postgres"
	TypeMongo          Type = "mongo"
	TypeMongoLegacy    Type = "mongo_legacy"
	TypeFirebirdLegacy Type = "firebird_legacy"
)

type GlobalConfiguration struct {
	Path                 string
	TmpPath              string
	ShExecutable         string
	PgdumpExecutable     string
	Mongodump5Executable string
	Mongodump4Executable string
	GbakExecutable       string
}

type Configuration struct {
	Type Type
	Name string

	//override destination path
	//when empty, path will be global path + dumper name
	Path string

	//override tmp path
	TmpPath string

	//variables to pass to dump executable
	Vars map[string]string

	//make daily dumps
	Daily bool
	Days  int

	//make weekly dumps
	Weekly bool
	Weeks  int

	//make monthly dumps
	Monthly bool
	Months  int
}
