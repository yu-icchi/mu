package archive

//go:generate mkdir -p mock
//go:generate mockgen -source=archive.go -package=mock -destination=mock/mock.go Archive

type Archive interface {
	Decompress(dest, src string) error
}
