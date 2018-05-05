// +build gofuzz

package admproto

func Fuzz(data []byte) int {
	var r Reader
	var err error

	data, _, _, err = r.Begin(data)
	if err != nil {
		return 0
	}

	for len(data) > 0 {
		data, _, _, err = r.Next(data)
		if err != nil {
			return 0
		}
	}

	return 1
}
