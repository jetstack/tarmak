package packer

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	logrus "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
)

const PackerTagEnvironment = "tarmak_environment"
const PackerTagBaseImageName = "tarmak_base_image_name"

type Packer struct {
	*tarmakDocker.App
	log    *logrus.Entry
	tarmak config.Tarmak

	imageID *string
}

func New(tarmak config.Tarmak) *Packer {
	log := tarmak.Log().WithField("module", "packer")

	app := tarmakDocker.NewApp(
		tarmak,
		log,
		"jetstack/tarmak-packer",
		"packer",
	)

	p := &Packer{
		App:    app,
		tarmak: tarmak,
		log:    log,
	}
	for key, val := range p.tags() {
		p.log = p.log.WithField(key, val)
	}

	return p
}

func (p *Packer) tags() map[string]string {
	return map[string]string{
		PackerTagEnvironment:   p.tarmak.Context().Environment().Name,
		PackerTagBaseImageName: p.tarmak.Context().BaseImage,
	}
}

func (p *Packer) QueryAMIID() (amiID string, err error) {
	if p.imageID != nil {
		return *p.imageID, nil
	}

	env := p.tarmak.Context().Environment()
	providerName := env.ProviderName()

	if providerName == config.ProviderNameAWS {
		p.log.Debug("querying AWS for latest matching AMI image")

		sess, err := env.AWS.Session()
		if err != nil {
			return "", err
		}

		svc := ec2.New(sess)

		filters := []*ec2.Filter{}
		for key, value := range p.tags() {
			filters = append(filters, &ec2.Filter{
				Name:   aws.String(fmt.Sprintf("tag:%s", key)),
				Values: []*string{aws.String(value)},
			})
		}

		images, err := svc.DescribeImages(&ec2.DescribeImagesInput{
			Filters: filters,
		})
		if err != nil {
			return "", err
		}

		if len(images.Images) == 0 {
			return "", fmt.Errorf("no image found, tags: %+v", p.tags())
		}

		var latest *ec2.Image
		var latestTime time.Time

		formatRFC3339aws := "2006-01-02T15:04:05.999Z07:00"

		for _, image := range images.Images {
			myTime, err := time.Parse(formatRFC3339aws, *image.CreationDate)
			if err != nil {
				return "", fmt.Errorf("error parsing time stamp: %s", err)
			}
			if latest == nil || myTime.After(latestTime) {
				latest = image
				latestTime = myTime
			}
		}

		p.log.Infof("found %d matching images, using latest: '%s'", len(images.Images), *latest.ImageId)

		p.imageID = latest.ImageId
		return *latest.ImageId, nil
	}

	return "", fmt.Errorf("unsupported provider: %s", providerName)
}

func (p *Packer) Build() (amiID string, err error) {
	c := p.Container()

	// set tarmak environment vars vars
	for key, value := range p.tags() {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	// get aws secrets
	if environmentProvider, err := p.tarmak.Context().ProviderEnvironment(); err != nil {
		return "", fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		c.Env = append(c.Env, environmentProvider...)
	}

	c.WorkingDir = "/packer"
	c.Cmd = []string{"sleep", "3600"}

	err = c.Prepare()
	if err != nil {
		return "", err
	}

	// make sure container get's cleaned up
	defer c.CleanUpSilent(p.log)

	buildSourcePath := filepath.Join(
		p.tarmak.RootPath(),
		"packer",
		fmt.Sprintf("%s.json", p.tarmak.Context().BaseImage),
	)

	buildContent, err := ioutil.ReadFile(buildSourcePath)
	if err != nil {
		return "", err
	}

	buildPath := "build.json"

	buildTar, err := tarmakDocker.TarStreamFromFile(buildPath, string(buildContent))
	if err != nil {
		return "", err
	}

	err = c.UploadToContainer(buildTar, "/packer")
	if err != nil {
		return "", err
	}
	p.log.Debug("copied packer build state")

	err = c.Start()
	if err != nil {
		return "", fmt.Errorf("error starting container: %s", err)
	}

	returnCode, err := c.Execute("packer", []string{"build", buildPath})
	if err != nil {
		return "", err
	}
	if exp, act := 0, returnCode; exp != act {
		return "", fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}

	return "unknown", nil
}
