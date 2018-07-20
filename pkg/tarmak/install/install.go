package install

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

const (
	terraformUrl        = "https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_amd64.zip"
	terraformHashDarwin = "6514a8fe5a344c5b8819c7f32745cd571f58092ffc9bbe9ea3639799b97ced5f"
	terraformHashLinux  = "6b8ce67647a59b2a3f70199c304abca0ddec0e49fd060944c26f666298e23418"
	terraformPath       = "/opt/tarmak/terraform"
	terraformVersion    = "0.11.7"

	packerUrl        = "https://releases.hashicorp.com/packer/%s/packer_%s_%s_amd64.zip"
	packerHashLinux  = "bc58aa3f3db380b76776e35f69662b49f3cf15cf80420fc81a15ce971430824c"
	packerHashDarwin = "3d546eff8179fc0de94ad736718fdaebdfc506536203eade732d9d218fbb347c"
	packerPath       = "/opt/tarmak/packer"
	packerVersion    = "1.2.5"

	tarmakDir = "/opt/tarmak"
)

type Install struct {
	tarmak interfaces.Tarmak
	log    *logrus.Entry
}

type dependency struct {
	path, zippath, url, hash string
	log                      *logrus.Entry
}

func New(t interfaces.Tarmak) *Install {
	return &Install{
		tarmak: t,
		log:    t.Log(),
	}
}

func (i *Install) Ensure() error {
	f, err := os.Stat(tarmakDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(tarmakDir, os.FileMode(0777)); err != nil {
				return err
			}

		} else {
			return err
		}
	} else {
		if !f.IsDir() {
			return fmt.Errorf("file '%s' is not a directory", tarmakDir)
		}
	}

	var result *multierror.Error
	for _, d := range i.dependencies() {
		f, err := os.Stat(d.path)
		if err != nil {
			if os.IsNotExist(err) {
				if err := d.downloadBinary(); err != nil {
					result = multierror.Append(result, err)
				}

			} else {
				result = multierror.Append(result, err)
			}

			continue
		}

		if f.IsDir() {
			err = fmt.Errorf("path '%s' is a directory", d.path)
			result = multierror.Append(result, err)
			continue
		}

		if (f.Mode() & os.FileMode(0766)) != 0 {
			err = fmt.Errorf("file '%s' does not match expected permissions(0766): %s", d.path, os.FileMode(0766))
			result = multierror.Append(result, err)
			continue
		}
	}

	if result != nil {
		return result.ErrorOrNil()
	}

	for _, d := range i.dependencies() {
		if err := d.checkSum(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (d *dependency) downloadBinary() error {
	d.log.Infof("Installing binary '%s' from '%s'", d.path, d.url)

	f, err := os.Create(d.zippath)
	defer f.Close()
	if err != nil {
		return err
	}

	resp, err := http.Get(d.url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	if err := d.unzip(); err != nil {
		return err
	}

	//if err := os.Remove(d.zippath); err != nil {
	//	return err
	//}

	return nil
}

func (d *dependency) checkSum() error {
	d.log.Infof("Checking binary checksum '%s'", d.path)

	f, err := os.Open(d.zippath)
	defer f.Close()
	if err != nil {
		return err
	}

	hash := sha256.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return err
	}

	s := hex.EncodeToString(hash.Sum(nil))

	if s != d.hash {
		return fmt.Errorf("binary at '%s' does not match expected hash", d.path)
	}

	d.log.Infof("'%s' Passed.", d.path)

	return nil
}

func (i *Install) dependencies() []*dependency {
	var linux bool
	var dependencies []*dependency

	if runtime.GOOS == "linux" {
		linux = true
	}

	terraform := &dependency{
		path:    terraformPath,
		zippath: fmt.Sprintf("%s.zip", terraformPath),
		url:     fmt.Sprintf(terraformUrl, terraformVersion, terraformVersion, runtime.GOOS),
		log:     i.log,
	}
	if linux {
		terraform.hash = terraformHashLinux
	} else {
		terraform.hash = terraformHashDarwin
	}
	dependencies = append(dependencies, terraform)

	packer := &dependency{
		path:    packerPath,
		zippath: fmt.Sprintf("%s.zip", packerPath),
		url:     fmt.Sprintf(packerUrl, packerVersion, packerVersion, runtime.GOOS),
		log:     i.log,
	}
	if linux {
		packer.hash = packerHashLinux
	} else {
		packer.hash = packerHashDarwin
	}
	dependencies = append(dependencies, packer)

	return dependencies
}

func (d *dependency) unzip() error {
	r, err := zip.OpenReader(d.zippath)
	if err != nil {
		return err
	}
	defer r.Close()

	if len(r.File) != 1 {
		return fmt.Errorf("got unexpected number of files from zip, exp=1 got=%d", len(r.File))
	}

	rc, err := r.File[0].Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	f, err := os.OpenFile(d.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, r.File[0].Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return err
	}

	return nil
}
