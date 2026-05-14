package generate

import (
	"fmt"
	"io"
	"strings"
)

const plistFragment = `<key>CFBundleURLTypes</key>
<array>
	<dict>
		<key>CFBundleURLName</key>
		<string>%s</string>
		<key>CFBundleURLSchemes</key>
		<array>
			<string>%s</string>
		</array>
	</dict>
</array>`

// PlistSnippet returns the CFBundleURLTypes XML fragment for use in Info.plist.
func PlistSnippet(spec HandlerSpec) (string, error) {
	id, err := spec.HandlerID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(plistFragment, id, spec.Scheme), nil
}

// PatchPlist reads an existing Info.plist and injects CFBundleURLTypes before </dict>\n</plist>.
func PatchPlist(r io.Reader, spec HandlerSpec) (string, error) {
	if err := spec.Validate(); err != nil {
		return "", err
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	snippet, _ := PlistSnippet(spec)
	src := string(b)
	if out := strings.Replace(src, "</dict>\n</plist>", snippet+"\n</dict>\n</plist>", 1); out != src {
		return out, nil
	}
	return strings.Replace(src, "</dict></plist>", snippet+"\n</dict></plist>", 1), nil
}
