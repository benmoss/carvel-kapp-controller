// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
//go:build tools

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"sync"

	"flag"

	"cuelang.org/go/cue/cuecontext"
)

//go:embed deps.cue
var dependencies string

type dependency struct {
	Name      string
	Version   string
	Dev       bool
	Artifacts map[string]map[string]artifact
}

type artifact struct {
	URL            string
	Checksum       *string
	TarballSubpath *string
}

func main() {
	targetOS := flag.String("os", runtime.GOOS, "os to install dependencies for")
	targetArch := flag.String("arch", runtime.GOARCH, "arch to install dependencies for")
	destDir := flag.String("dest", "", "directory to which binaries will be installed")
	dev := flag.Bool("dev", false, "only install dev dependencies (default false)")
	flag.Parse()

	if *destDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	c := cuecontext.New()
	v := c.CompileString(dependencies)
	var deps map[string]dependency
	if err := v.Decode(&deps); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, dep := range deps {
		if *dev && !dep.Dev {
			continue
		}
		wg.Add(1)
		go func(dep dependency) {
			defer wg.Done()
			downloadAndVerify(dep, *targetOS, *targetArch, *destDir)
		}(dep)
	}
	wg.Wait()
}

func downloadAndVerify(dep dependency, targetOS, targetArch, destDir string) {
	artifact := dep.Artifacts[targetOS][targetArch]
	if artifact.Checksum == nil {
		log.Printf("%s not supported on platform %s/%s", dep.Name, targetOS, targetArch)
		return
	}

	resp, err := http.Get(artifact.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("bad status code retrieving file: %s: %s", dep.Name, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	hasher := sha256.New()
	_, err = hasher.Write(content)
	if err != nil {
		log.Fatal(err)
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != *artifact.Checksum {
		log.Fatalf("%s: wrong checksum, expected: %s, got: %s", dep.Name, *artifact.Checksum, actual)
	}

	if artifact.TarballSubpath != nil {
		gzReader, err := gzip.NewReader(bytes.NewReader(content))
		if err != nil {
			log.Fatal(err)
		}
		tgzReader := tar.NewReader(gzReader)
		for {
			hdr, err := tgzReader.Next()
			if err == io.EOF {
				log.Fatal("file not found")
			}
			if hdr.Name == *artifact.TarballSubpath {
				out, err := os.Create(path.Join(destDir, dep.Name))
				if err != nil {
					log.Fatal(err)
				}
				_, err = io.Copy(out, tgzReader)
				if err != nil {
					log.Fatal(err)
				}
				if err := out.Chmod(0777); err != nil {
					log.Fatal(err)
				}
				break
			}
		}
	} else {
		if err = os.WriteFile(path.Join(destDir, dep.Name), content, 0777); err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("%s downloaded", dep.Name)
}
