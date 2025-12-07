package constant

type ImportStatus string

const (
	ImportStatusSuccess ImportStatus = "SUCCESS"
	ImportStatusFailed  ImportStatus = "FAILED"
)

func (s ImportStatus) String() string {
	return string(s)
}
