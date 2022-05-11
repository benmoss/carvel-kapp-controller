// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/yaml"
)

var (
	cfgFile      string
	dependencies []*dependency
	client       *http.Client
)

type dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// Template for the download URL of the artifact
	URLTemplate string `json:"urlTemplate"`
	// Whether this should be installed by the --dev flag
	Dev bool `json:"dev"`
	// Whether this should be installed without the --dev flag
	Prod bool `json:"prod"`
	// Optional ways of configuring how we check for updates
	AutoUpdate *autoUpdate `json:"autoupdate,omitempty"`
	// map["linux"]map["amd64"] => "sha256"
	Checksums      map[string]map[string]string `json:"checksums"`
	TarballSubpath *string                      `json:"tarballSubpath,omitempty"`
}

type autoUpdate struct {
	Github    string `json:"github"`
	Checksums struct {
		File         *string `json:"file,omitempty"`
		ReleaseNotes *bool   `json:"releaseNotes,omitempty"`
	} `json:"checksums,omitempty"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string
		BrowserDownloadURL string `json:"browser_download_url"`
	}
	Body string `json:"body"`
}

type fields struct {
	Name,
	Version,
	OS,
	Arch string
}

type installCommand struct {
	*cobra.Command
	arch, os, destDir string
	dev               bool
}

func newInstallCommand() *cobra.Command {
	c := &installCommand{}
	command := &cobra.Command{
		Use:   "install",
		Short: "install dependencies",
		Run: func(_ *cobra.Command, _ []string) {
			if err := c.Install(); err != nil {
				log.Fatal(err)
			}
		},
	}
	command.Flags().StringVar(&c.os, "os", runtime.GOOS, "os to install dependencies for")
	command.Flags().StringVar(&c.arch, "arch", runtime.GOARCH, "arch to install dependencies for")
	command.Flags().StringVarP(&c.destDir, "destination", "d", "", "directory to which binaries will be installed (required)")
	command.Flags().BoolVar(&c.dev, "dev", false, "only install dev dependencies (default false)")
	command.MarkFlagRequired("destination")
	return command
}

func (c *installCommand) Install() error {
	g, ctx := errgroup.WithContext(context.Background())
	for _, dep := range dependencies {
		if c.dev && !dep.Dev {
			continue
		} else if !c.dev && !dep.Prod {
			continue
		}
		dep := dep // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			if err := c.downloadAndVerify(ctx, dep); err != nil {
				return fmt.Errorf("downloading %s: %w", dep.Name, err)
			}
			return nil
		})
	}
	return g.Wait()
}

// downloads the dependency by its url template, verifies the checksum, and
// optionally extracts the executable from the tarball
func (c *installCommand) downloadAndVerify(ctx context.Context, dep *dependency) error {
	arches, ok := dep.Checksums[c.os]
	if !ok {
		return fmt.Errorf("%s not supported on os %s", dep.Name, c.os)
	}
	checksum, ok := arches[c.arch]
	if !ok {
		return fmt.Errorf("%s not supported on platform %s/%s", dep.Name, c.os, c.arch)
	}

	urlTpl, err := template.New(dep.Name).Parse(dep.URLTemplate)
	if err != nil {
		return err
	}
	var url bytes.Buffer
	fields := fields{Name: dep.Name, Version: dep.Version, OS: c.os, Arch: c.arch}
	if err := urlTpl.Execute(&url, fields); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code retrieving url: %s: %s", url.String(), resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("%s downloaded", dep.Name)

	hasher := sha256.New()
	_, err = hasher.Write(content)
	if err != nil {
		return err
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != checksum {
		return fmt.Errorf("%s: wrong checksum, expected: %s, got: %s", dep.Name, checksum, actual)
	}
	log.Printf("%s validated", dep.Name)

	if dep.TarballSubpath != nil {
		tarballTpl, err := template.New(dep.Name).Parse(*dep.TarballSubpath)
		if err != nil {
			return err
		}
		var tarballSubpath bytes.Buffer
		if err := tarballTpl.Execute(&tarballSubpath, fields); err != nil {
			return err
		}
		gzReader, err := gzip.NewReader(bytes.NewReader(content))
		if err != nil {
			return err
		}
		tgzReader := tar.NewReader(gzReader)
		for {
			hdr, err := tgzReader.Next()
			if err == io.EOF {
				return fmt.Errorf("tarball subpath not found: %s", tarballSubpath.String())
			}
			if hdr.Name == tarballSubpath.String() {
				out, err := os.Create(path.Join(c.destDir, dep.Name))
				if err != nil {
					return err
				}
				_, err = io.Copy(out, tgzReader)
				if err != nil {
					return err
				}
				if err := out.Chmod(0777); err != nil {
					return err
				}
				log.Printf("%s extracted", dep.Name)
				break
			}
		}
	} else {
		if err = os.WriteFile(path.Join(c.destDir, dep.Name), content, 0777); err != nil {
			return err
		}
	}
	return nil
}

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "update dependencies",
		PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
			depsYAML, err := yaml.Marshal(&dependencies)
			if err != nil {
				return err
			}
			err = os.WriteFile(cfgFile, depsYAML, 0644)
			if err != nil {
				return err
			}
			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			g, ctx := errgroup.WithContext(context.Background())
			for _, dep := range dependencies {
				if dep.AutoUpdate == nil {
					continue
				}
				dep := dep // https://golang.org/doc/faq#closures_and_goroutines
				g.Go(func() error {
					if err := update(ctx, dep); err != nil {
						return fmt.Errorf("updating %s: %w", dep.Name, err)
					}
					return nil
				})
			}
			if err := g.Wait(); err != nil {
				log.Fatal(err)
			}
		},
	}
}

// updates the version of the dependency to the one GitHub considers latest and updates checksums
func update(ctx context.Context, dep *dependency) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", dep.AutoUpdate.Github)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code from server: %s", resp.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}

	// ParseTolerant allows the "v" prefix in the version string
	current, err := semver.ParseTolerant(dep.Version)
	if err != nil {
		return fmt.Errorf("err parsing semver from version %q: %w", dep.Version, err)
	}
	latest, err := semver.ParseTolerant(release.TagName)
	if err != nil {
		return fmt.Errorf("err parsing semver from version %q: %w", release.TagName, err)
	}

	// short-circuit if we're already at latest
	if latest.LTE(current) {
		log.Printf("%s is already at latest version %s", dep.Name, latest.String())
		return nil
	}
	log.Printf("Updating %s to %s", dep.Name, latest.String())

	// add the "v" prefix back
	dep.Version = "v" + latest.String()

	return updateChecksums(ctx, dep, &release)
}

func newSyncChecksums() *cobra.Command {
	return &cobra.Command{
		Use:   "sync-checksums",
		Short: "sync checksums for the current version",
		RunE: func(_ *cobra.Command, _ []string) error {
			g, ctx := errgroup.WithContext(context.Background())
			for _, dep := range dependencies {
				if dep.AutoUpdate == nil {
					continue
				}
				dep := dep // https://golang.org/doc/faq#closures_and_goroutines
				g.Go(func() error {
					release, err := getRelease(ctx, dep)
					if err != nil {
						return fmt.Errorf("getting release %s: %w", dep.Name, err)
					}
					if err := updateChecksums(ctx, dep, release); err != nil {
						return fmt.Errorf("updating %s: %w", dep.Name, err)
					}
					return nil
				})
			}
			return g.Wait()
		},
	}
}

// get the release specified by the current version of this dependency
func getRelease(ctx context.Context, dep *dependency) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", dep.AutoUpdate.Github, dep.Version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code from server fetching %s: %s", url, resp.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// update the checksums of this dependency to those in the github release
func updateChecksums(ctx context.Context, dep *dependency, release *githubRelease) error {
	urlTpl, err := template.New(dep.Name).Parse(dep.URLTemplate)
	if err != nil {
		return err
	}
	// map of filenames => checksums
	checksums := map[string]string{}
	// get checksums from a checksums.txt file
	if dep.AutoUpdate.Checksums.File != nil {
		var checksumFile []byte
		// look for the checksum file in the release assets
		for _, asset := range release.Assets {
			if asset.Name == *dep.AutoUpdate.Checksums.File {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
				if err != nil {
					return err
				}
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				bs, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				checksumFile = bs
				break
			}
		}
		if len(checksumFile) == 0 {
			return fmt.Errorf("did not find checksums file in release assets")
		}
		// extract checksums to the map
		scanner := bufio.NewScanner(bytes.NewReader(checksumFile))
		for scanner.Scan() {
			split := strings.Fields(scanner.Text())
			if len(split) != 2 {
				return fmt.Errorf("expected checksums file to be in format 'sha256 filename'")
			}
			filename := strings.TrimLeft(split[1], "./")
			checksums[filename] = split[0]
		}
	}

	// get checksums from release notes: https://regex101.com/r/8DbUbt/1
	if dep.AutoUpdate.Checksums.ReleaseNotes != nil && *dep.AutoUpdate.Checksums.ReleaseNotes {
		regex := regexp.MustCompile(`(?m)([a-z0-9]{64})[\s./]+([a-zA-Z0-9-.]*)`)
		allMatches := regex.FindAllStringSubmatch(release.Body, -1)
		for _, matches := range allMatches {
			if len(matches) != 3 {
				// full match, checksum, filename
				log.Fatalf("expected 3 matches in release notes: %v", matches)
			}
			checksums[matches[2]] = matches[1]
		}
	}
	// TODO: figure out how to get checksums for helm, sops, age

	for os, arches := range dep.Checksums {
		for arch := range arches {
			var url bytes.Buffer
			// generate the URL
			if err := urlTpl.Execute(&url, fields{Name: dep.Name, Version: dep.Version, OS: os, Arch: arch}); err != nil {
				log.Fatal(err)
			}
			// get just the filename
			filename := path.Base(url.String())
			// find the checksum for this file
			checksum, ok := checksums[filename]
			if !ok {
				return fmt.Errorf("no checksum found for filename %s", filename)
			}
			// update the checksum in the dependency
			arches[arch] = checksum
		}
	}
	return nil
}

type githubTransport string

func newGithubTransport() http.RoundTripper {
	token := os.Getenv("GITHUB_TOKEN")
	return githubTransport(token)
}

func (t githubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer: %s", t))
	return http.DefaultTransport.RoundTrip(r)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	client = &http.Client{
		// try to use a Github API token from the environment to avoid rate-limiting
		Transport: newGithubTransport(),
	}
	var rootCmd = &cobra.Command{Use: "dependencies"}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "hack/dependencies.yml", "config file")
	cobra.OnInitialize(func() {
		depsYAML, err := os.ReadFile(cfgFile)
		if err != nil {
			log.Fatal(err)
		}
		if err := yaml.Unmarshal(depsYAML, &dependencies); err != nil {
			log.Fatal(err)
		}
	})

	update := newUpdateCommand()
	update.AddCommand(newSyncChecksums())
	rootCmd.AddCommand(newInstallCommand(), update)
	rootCmd.Execute()
}
