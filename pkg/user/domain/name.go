package domain

const (
	// NameResultAvailable defines successful name check or change result.
	NameResultAvailable int32 = 0
	// NameResultTaken defines already-taken username result.
	NameResultTaken int32 = 1
	// NameResultInvalid defines invalid username format result.
	NameResultInvalid int32 = 2
	// NameResultNotAllowed defines blocked rename result.
	NameResultNotAllowed int32 = 3
)

// NameResult defines one username check/change result payload.
type NameResult struct {
	// ResultCode stores protocol-compatible result code.
	ResultCode int32
	// Name stores requested username value.
	Name string
	// Suggestions stores fallback username suggestions.
	Suggestions []string
}
