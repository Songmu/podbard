// Code generated by "stringer -type=MediaType"; DO NOT EDIT.

package cast

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MP3-1]
	_ = x[M4A-2]
}

const _MediaType_name = "MP3M4A"

var _MediaType_index = [...]uint8{0, 3, 6}

func (i MediaType) String() string {
	i -= 1
	if i < 0 || i >= MediaType(len(_MediaType_index)-1) {
		return "MediaType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _MediaType_name[_MediaType_index[i]:_MediaType_index[i+1]]
}
