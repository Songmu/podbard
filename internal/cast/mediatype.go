package cast

type MediaType int

//go:generate go run golang.org/x/tools/cmd/stringer -type=MediaType
const (
	MP3 MediaType = iota + 1
	M4A
)

var mediaTypeExtMap = map[string]MediaType{
	".mp3": MP3,
	".m4a": M4A,
}

func GetMediaTypeByExt(ext string) (MediaType, bool) {
	mt, ok := mediaTypeExtMap[ext]
	return mt, ok
}

func IsSupportedMediaExt(ext string) bool {
	_, ok := mediaTypeExtMap[ext]
	return ok
}
