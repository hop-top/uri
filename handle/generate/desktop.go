package generate

import (
	"bytes"
	"text/template"
)

var desktopTmpl = template.Must(template.New("desktop").Parse(`[Desktop Entry]
Type=Application
Name={{.Name}}
Exec={{.Exec}} %u
MimeType=x-scheme-handler/{{.Scheme}};
NoDisplay=true
X-Hop-Handler-ID={{.HandlerID}}
`))

// DesktopFile returns a .desktop file for registering the given URL scheme on Linux.
func DesktopFile(spec HandlerSpec) (string, error) {
	id, err := spec.HandlerID()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = desktopTmpl.Execute(&buf, struct {
		Scheme    string
		Exec      string
		Name      string
		HandlerID string
	}{
		Scheme:    spec.Scheme,
		Exec:      spec.AppPath,
		Name:      spec.displayName(),
		HandlerID: id,
	})
	return buf.String(), err
}

// DesktopFilename returns a unique .desktop filename for the handler artifact.
func DesktopFilename(spec HandlerSpec) (string, error) {
	id, err := spec.HandlerID()
	if err != nil {
		return "", err
	}
	return id + ".desktop", nil
}
