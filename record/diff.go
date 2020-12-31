package record

import (
	"bytes"
	"github.com/kr/binarydist"
)

type patch []byte
type frame []byte

func doDiff(prev, next []byte) (err error, p patch) {
	oldReader := bytes.NewReader(prev)
	newReader := bytes.NewReader(next)
	writer := new(bytes.Buffer)

	if err = binarydist.Diff(oldReader, newReader, writer); err != nil {
		return
	}

	return nil, patch(writer.Bytes())
}

func doPatch(cur []byte, p patch) (err error, f frame) {
	oldReader := bytes.NewReader(cur)
	patchReader := bytes.NewReader(p)
	newWriter := new(bytes.Buffer)

	if err = binarydist.Patch(oldReader, newWriter, patchReader); err != nil {
		return
	}

	return nil, newWriter.Bytes()
}