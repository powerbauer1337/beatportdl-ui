//go:build notag
// +build notag

package taglib

import "errors"

var (
	ErrInvalid   = errors.New("invalid file")
	ErrStripMp4  = errors.New("cannot strip mp4 tags")
	ErrSave      = errors.New("cannot save file")
	ErrNoPicture = errors.New("no picture")
	ErrNoTagging = errors.New("tagging functionality is disabled in this build")
)

type File struct{}

func Read(filename string) (*File, error) {
	return nil, ErrNoTagging
}

func (f *File) Close() {}

func (f *File) Save() error {
	return ErrNoTagging
}

func (f *File) SetItemMp4(key, value string) {}

func (f *File) StripMp4() error {
	return ErrNoTagging
}

func (f *File) GetProperty(property string) string {
	return ""
}

func (f *File) SetProperty(property string, value *string) {}

func (f *File) PropertyKeys() ([]string, error) {
	return nil, ErrNoTagging
}

func (f *File) SampleRate() int {
	return 0
}

type Picture struct {
	MimeType    string
	PictureType string
	Description string
	Data        []byte
	Size        uint
}

func (f *File) GetPicture() (*Picture, error) {
	return nil, ErrNoTagging
}

func (f *File) SetPicture(picture *Picture) error {
	return ErrNoTagging
}

func (f *File) ComplexPropertyKeys() ([]string, error) {
	return nil, ErrNoTagging
}
