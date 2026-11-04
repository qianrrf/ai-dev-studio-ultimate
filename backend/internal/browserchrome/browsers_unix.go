//go:build darwin || linux

package browserchrome

import "path/filepath"

func browsersUnder(base string, channels ...[2]string) []Browser {
	out := make([]Browser, 0, len(channels))
	for _, channel := range channels {
		out = append(out, Browser{Name: channel[0], UserDataDir: filepath.Join(base, filepath.FromSlash(channel[1]))})
	}
	return out
}
