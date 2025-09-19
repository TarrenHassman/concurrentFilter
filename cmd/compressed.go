package cmd

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"
)

const numOfZipWorkers = 10

type entry struct {
	name string
	rc   io.ReadCloser
}

func compressedFilter() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	entCh := make(chan entry, numOfZipWorkers)
	zpathCh := make(chan string, numOfZipWorkers)

	group, ctx := errgroup.WithContext(context.Background())

	for i := 0; i < numOfZipWorkers; i++ {
		group.Go(func() error {
			return zipWorker(ctx, entCh, zpathCh)
		})
	}

	group.Go(func() error {
		defer close(entCh) // Signal workers to stop.
		return entryProvider(ctx, entCh)
	})

	err := group.Wait()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("output.zip", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	zw := zip.NewWriter(f)

	close(zpathCh)
	for path := range zpathCh {
		zrd, err := zip.OpenReader(path)
		if err != nil {
			log.Fatal(err)
		}
		for _, zf := range zrd.File {
			err := zw.Copy(zf)
			if err != nil {
				log.Fatal(err)
			}
		}
		_ = zrd.Close()
		_ = os.Remove(path)
	}
	err = zw.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func entryProvider(ctx context.Context, entCh chan<- entry) error {
	for i := 0; i < 2*numOfZipWorkers; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entCh <- entry{
			name: fmt.Sprintf("file_%d", i+1),
			rc:   io.NopCloser(strings.NewReader(fmt.Sprintf("content %d", i+1))),
		}:
		}
	}
	return nil
}

func zipWorker(ctx context.Context, entCh <-chan entry, zpathch chan<- string) error {
	f, err := os.CreateTemp(".", "tmp-part-*")
	if err != nil {
		return err
	}

	zw := zip.NewWriter(f)
Loop:
	for {
		var (
			ent entry
			ok  bool
		)
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break Loop
		case ent, ok = <-entCh:
			if !ok {
				break Loop
			}
		}

		hdr := &zip.FileHeader{
			Name:   ent.name,
			Method: zip.Deflate, // zip.Store can also be used.
		}
		hdr.SetMode(0644)

		w, e := zw.CreateHeader(hdr)
		if e != nil {
			_ = ent.rc.Close()
			err = e
			break
		}

		_, e = io.Copy(w, ent.rc)
		_ = ent.rc.Close()
		if e != nil {
			err = e
			break
		}
	}

	if e := zw.Close(); e != nil && err == nil {
		err = e
	}
	if e := f.Close(); e != nil && err == nil {
		err = e
	}
	if err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case zpathch <- f.Name():
		}
	}
	return err
}
