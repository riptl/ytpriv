package api

type Err int
const (
	GenericError = Err(iota)
	VideoUnavailable
)

func (e Err) Error() string { switch e {
	case VideoUnavailable:
		return "video unavailable"
	default:
		return "unknown error"
}}
