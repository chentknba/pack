package lua

const (
	TAB_1   = "\t"
	TAB_2   = "\t\t"
	TAB_10  = "\t\t\t\t\t\t\t\t\t\t"
	COMMENT = "--"
	AUTO_FILE_DESC = "-- auto generated, modification is not permitted."
)

type result struct {
	file, content string
}