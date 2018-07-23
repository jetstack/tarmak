package install

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

const (
	terraformVersion       = "0.11.7"
	terraformUrl           = "https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_amd64.zip"
	terraformHashLinuxZip  = "6b8ce67647a59b2a3f70199c304abca0ddec0e49fd060944c26f666298e23418"
	terraformHashDarwinZip = "6514a8fe5a344c5b8819c7f32745cd571f58092ffc9bbe9ea3639799b97ced5f"
	terraformHashLinux     = "00cc2e727e662fb81c789b2b8371d82d6be203ddc76c49232ed9c17b4980949a"
	terraformHashDarwin    = "e8460143408184b383baba6226d4076887aa774e522b19c84f2d65070c1a1430"

	packerVersion       = "1.2.5"
	packerUrl           = "https://releases.hashicorp.com/packer/%s/packer_%s_%s_amd64.zip"
	packerHashLinuxZip  = "bc58aa3f3db380b76776e35f69662b49f3cf15cf80420fc81a15ce971430824c"
	packerHashDarwinZip = "3d546eff8179fc0de94ad736718fdaebdfc506536203eade732d9d218fbb347c"
	packerHashLinux     = "fd9d6c7acdeacfd1a08487a1f3308269f7d01f64950158133f6f0f438d3d1902"
	packerHashDarwin    = "fa89d4e1ab14cd934d560d5efb5a886ed9003ad075b3bc514dd703a5db1ef1fb"
)

type Install struct {
	tarmak interfaces.Tarmak
	log    *logrus.Entry
}

type dependency struct {
	path, zippath, url string
	hash, ziphash      string
	log                *logrus.Entry
}

func New(t interfaces.Tarmak) *Install {
	return &Install{
		tarmak: t,
		log:    t.Log(),
	}
}

func (i *Install) Ensure() error {
	var result *multierror.Error
	for _, d := range i.dependencies() {
		f, err := os.Stat(d.path)
		if err != nil {
			if os.IsNotExist(err) {
				if err := d.downloadBinary(); err != nil {
					result = multierror.Append(result, err)
				}

				f, err = os.Stat(d.path)
				if err != nil {
					result = multierror.Append(result, err)
					continue
				}

			} else {
				result = multierror.Append(result, err)

				continue
			}
		}

		if f.IsDir() {
			err = fmt.Errorf("path '%s' is a directory", d.path)
			result = multierror.Append(result, err)
			continue
		}

		if (f.Mode() & os.FileMode(0022)) != 0 {
			err = fmt.Errorf("file '%s' does not match expected permissions(0755): %s", d.path, os.FileMode(0755))
			result = multierror.Append(result, err)
			continue
		}
	}

	if result.ErrorOrNil() != nil {
		return result.ErrorOrNil()
	}

	for _, d := range i.dependencies() {
		if err := d.checkSum(d.path, d.hash); err != nil {
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

	if err := d.checkSum(d.zippath, d.ziphash); err != nil {
		return err
	}

	if err := d.unzip(); err != nil {
		return err
	}

	return nil
}

func (d *dependency) checkSum(path, hashStr string) error {
	f, err := os.Open(path)
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

	if s != hashStr {
		return fmt.Errorf("file at '%s' does not match expected hash", path)
	}

	return nil
}

func (i *Install) dependencies() []*dependency {
	var linux bool
	var dependencies []*dependency

	if runtime.GOOS == "linux" {
		linux = true
	}

	path := filepath.Join(i.tarmak.ConfigPath(), "terraform")
	terraform := &dependency{
		path:    path,
		zippath: fmt.Sprintf("%s.zip", path),
		url:     fmt.Sprintf(terraformUrl, terraformVersion, terraformVersion, runtime.GOOS),
		log:     i.log,
	}
	if linux {
		terraform.hash = terraformHashLinux
		terraform.ziphash = terraformHashLinuxZip
	} else {
		terraform.hash = terraformHashDarwin
		terraform.ziphash = terraformHashDarwinZip
	}
	dependencies = append(dependencies, terraform)

	path = filepath.Join(i.tarmak.ConfigPath(), "packer")
	packer := &dependency{
		path:    path,
		zippath: fmt.Sprintf("%s.zip", path),
		url:     fmt.Sprintf(packerUrl, packerVersion, packerVersion, runtime.GOOS),
		log:     i.log,
	}
	if linux {
		packer.hash = packerHashLinux
		packer.ziphash = packerHashLinuxZip
	} else {
		packer.hash = packerHashDarwin
		packer.ziphash = packerHashDarwinZip
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
