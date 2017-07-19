package cmd

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/docker/docker/pkg/archive"
	"github.com/spf13/cobra"
)

var cfgFile string

const s3Prefix = "s3://"

func getReader(input string) (io.ReadCloser, error) {

	if strings.HasPrefix(input, s3Prefix) {
		pathParts := strings.Split(input[len(s3Prefix):len(input)], "/")
		bucket := pathParts[0]
		key := strings.Join(pathParts[1:len(pathParts)], "/")

		cfg := aws.NewConfig()
		awsSession := session.New(cfg)
		s3Service := s3.New(awsSession)

		result, err := s3Service.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return nil, fmt.Errorf("error getting s3 object %s: %s", input, err)
		}

		return result.Body, nil

	}

	f, err := os.Open(input)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %s", input, err)
	}
	return f, nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "wing",
	Short: "Wing is the agent that runs on every instance of Tarmak",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.DebugLevel)

		input := "puppet.tar.gz"
		if len(args) > 0 {
			input = args[0]
		}

		reader, err := getReader(input)
		if err != nil {
			log.Fatal(err)
		}

		tarReader, err := gzip.NewReader(reader)
		if err != nil {
			log.Fatal(err)
		}

		dir, err := ioutil.TempDir("", "tarmak-apply")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir) // clean up

		err = archive.Unpack(tarReader, dir, &archive.TarOptions{})
		if err != nil {
			log.Fatal(err)
		}
		tarReader.Close()
		reader.Close()

		puppetCmd := exec.Command(
			"puppet",
			"apply",
			"--environment",
			"production",
			"--hiera_config",
			filepath.Join(dir, "hiera.yaml"),
			"--modulepath",
			filepath.Join(dir, "modules"),
			filepath.Join(dir, "manifests/site.pp"),
		)

		stdoutPipe, err := puppetCmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		stderrPipe, err := puppetCmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}

		stdoutScanner := bufio.NewScanner(stdoutPipe)
		go func() {
			for stdoutScanner.Scan() {
				log.WithField("cmd", "puppet").Debug(stdoutScanner.Text())
			}
		}()

		stderrScanner := bufio.NewScanner(stderrPipe)
		go func() {
			for stderrScanner.Scan() {
				log.WithField("cmd", "puppet").Debug(stderrScanner.Text())
			}
		}()

		err = puppetCmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Waiting for command to finish...")
		err = puppetCmd.Wait()
		log.Printf("Command finished with error: %v", err)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
