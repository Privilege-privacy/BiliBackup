package core

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
)

const defaultChunkSize = 2 * 1024 * 1024

type Downloader struct {
	Process       int64
	FileName      string
	Url           string
	ContentLength uint64
	Headers       []*Header
	Client        *http.Client
	Sagements     []*Segement
}

type Header struct {
	Key, Value string
}

type Segement struct {
	Start, End uint64
}

func (s *Segement) getRange() string {
	return fmt.Sprintf("bytes=%d-%d", s.Start, s.End)
}

type Offset struct {
	io.WriterAt
	n int64
}

func (o *Offset) Write(b []byte) (n int, err error) {
	n, err = o.WriteAt(b, o.n)
	o.n += int64(n)
	return
}

func NewDownloader(url, fileName string, process int64) *Downloader {
	return &Downloader{
		Process:   process,
		FileName:  fileName,
		Url:       url,
		Headers:   []*Header{},
		Client:    &http.Client{},
		Sagements: []*Segement{},
	}
}

func (d *Downloader) init() error {
	client := resty.New().SetRetryCount(3).R()
	for _, v := range d.Headers {
		client.SetHeader(v.Key, v.Value)
	}
	resp, err := client.Get(d.Url)
	if err != nil {
		return err
	}
	defer resp.RawResponse.Body.Close()

	d.ContentLength = uint64(resp.RawResponse.ContentLength)

	process := uint64(d.Process)
	length := uint64(d.ContentLength)
	chunksize := length / process

	if chunksize >= 102400000 {
		chunksize /= 2
	} else if chunksize < defaultChunkSize {
		chunksize = defaultChunkSize
	}

	if chunksize*process > length {
		chunksize = length / process
	}

	chunkslen := length / chunksize
	d.Sagements = make([]*Segement, chunkslen)

	for i := uint64(0); i < chunkslen; i++ {
		sagement := &Segement{
			Start: i * chunksize,
			End:   (i+1)*chunksize - 1,
		}
		if i == chunkslen-1 {
			sagement.End = length - 1
		}
		d.Sagements[i] = sagement
	}
	return nil
}

func (d *Downloader) getRequest() (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, d.Url, nil)
	if err != nil {
		return nil, err
	}

	for _, v := range d.Headers {
		req.Header.Set(v.Key, v.Value)
	}

	return req, nil
}

func (d *Downloader) Write(b []byte) (n int, err error) {
	return n, nil
}

func (d *Downloader) downloadSegment(segemen *Segement, writer io.Writer) error {
	req, err := d.getRequest()
	if err != nil {
		return err
	}
	req.Header.Set("Range", segemen.getRange())
	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	expectedLength := int64(segemen.End - segemen.Start + 1)
	if resp.ContentLength != expectedLength {
		return fmt.Errorf("unexpected content length: expected %d, got %d", expectedLength, resp.ContentLength)
	}

	_, err = io.CopyN(writer, io.TeeReader(resp.Body, d), resp.ContentLength)
	return err
}

func (d *Downloader) Start() (err error) {
	var file *os.File

	if err := d.init(); err != nil {
		return err
	}

	color.Yellow("开始下载: %s\n", d.FileName)
	if file, err = os.Create(d.FileName); err != nil {
		return err
	}
	defer file.Close()

	if err = file.Truncate(int64(d.ContentLength)); err != nil {
		return err
	}

	var eg errgroup.Group
	for i := 0; i < len(d.Sagements); i++ {
		Index := i
		eg.Go(func() error {
			return d.downloadSegment(d.Sagements[Index], &Offset{file, int64(d.Sagements[Index].Start)})
		})
	}

	if err := eg.Wait(); err != nil {
		os.Remove(d.FileName)
		return err
	}
	return nil
}
