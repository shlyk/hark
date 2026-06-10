package notify

import (
	"encoding/json"
	"strings"
)

// audioProfile mirrors the parts of `system_profiler SPAudioDataType -json`
// output needed to find the default output device.
type audioProfile struct {
	SPAudioDataType []struct {
		Items []audioDevice `json:"_items"`
	} `json:"SPAudioDataType"`
}

type audioDevice struct {
	Name          string `json:"_name"`
	DefaultOutput string `json:"coreaudio_default_audio_output_device"`
	Transport     string `json:"coreaudio_device_transport"`
}

// HeadphonesConnected reports whether the default audio output looks like
// headphones: a Bluetooth device (AirPods etc.) or the built-in jack
// ("External Headphones"). Any detection failure returns false, so callers
// degrade to a silent banner rather than erroring.
func HeadphonesConnected(e Execer) bool {
	out, err := e.Output("system_profiler", "SPAudioDataType", "-json")
	if err != nil {
		return false
	}
	var profile audioProfile
	if json.Unmarshal(out, &profile) != nil {
		return false
	}
	for _, group := range profile.SPAudioDataType {
		for _, dev := range group.Items {
			if dev.DefaultOutput != "spaudio_yes" {
				continue
			}
			name := strings.ToLower(dev.Name)
			return strings.Contains(dev.Transport, "bluetooth") ||
				strings.Contains(name, "headphone") ||
				strings.Contains(name, "airpod")
		}
	}
	return false
}
