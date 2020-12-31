package record

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

var EOF = errors.New("EOF")

type Record struct {
	sync.Mutex
	offset  int
	limit   int
	current []byte
	Head    []byte  `json:"head"`
	Patches []patch `json:"states"`
}

func New() (*Record, error) {
	return &Record{
		current: nil,
		offset:  0,
		limit:   0,
		Head:    nil,
		Patches: make([]patch, 0),
	}, nil
}

func Load(fileName string) (*Record, error) {
	compressed, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %s", fileName, err)
	}

	decompress, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("error while reading gzip file %s: %s", fileName, err)
	}
	defer decompress.Close()

	raw, err := ioutil.ReadAll(decompress)
	if err != nil {
		return nil, fmt.Errorf("error while decompressing %s: %s", fileName, err)
	}

	rec := Record{}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err = decoder.Decode(&rec); err != nil {
		return nil, fmt.Errorf("error while decoding %s: %s", fileName, err)
	}

	rec.Reset()

	return &rec, nil
}

func (r *Record) Reset() {
	r.Lock()
	defer r.Unlock()
	r.limit = len(r.Patches)
	r.current = nil
	r.offset = 0
}

func (r *Record) Next(v interface{}) error {
	r.Lock()
	defer r.Unlock()

	var err error
	var raw []byte

	if r.current == nil {
		raw = r.Head
	} else if r.offset < r.limit {
		patch := r.Patches[r.offset]
		r.offset++
		if err, raw = doPatch(r.current, patch); err != nil {
			return err
		}

	} else {
		return EOF
	}

	r.current = raw

	return json.Unmarshal(raw, v)
}

func (r *Record) Add(v interface{}) error {
	r.Lock()
	defer r.Unlock()

	raw, err := json.Marshal(v)
	if err != nil {
		return err
	} else if r.Head == nil {
		r.Head = raw
	} else if err, patch := doDiff(r.current, raw); err != nil {
		return err
	} else {
		r.Patches = append(r.Patches, patch)
	}

	r.current = raw

	return nil
}

func (r *Record) Save(fileName string) error {
	r.Lock()
	defer r.Unlock()

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(r); err != nil {
		return err
	}

	data := buf.Bytes()
	compressed := new(bytes.Buffer)
	compress := gzip.NewWriter(compressed)

	if _, err := compress.Write(data); err != nil {
		return err
	} else if err = compress.Flush(); err != nil {
		return err
	} else if err = compress.Close(); err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, compressed.Bytes(), os.ModePerm)
}
