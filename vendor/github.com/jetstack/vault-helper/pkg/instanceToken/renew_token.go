package instanceToken

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

func (i *InstanceToken) TokenFromFile(path string) (token string, err error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	token = strings.TrimSpace(string(dat))

	return token, nil
}

func (i *InstanceToken) fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (i *InstanceToken) TokenRetrieve() (token string, err error) {
	exists, err := i.fileExists(i.TokenFilePath())
	if err != nil {
		return "", fmt.Errorf("error checking file exists: %v", err)
	}

	if exists {
		i.Log.Debugf("File exists: %s", i.TokenFilePath())
		token, err := i.TokenFromFile(i.TokenFilePath())
		if err != nil {
			return "", err
		}
		return token, nil
	}

	return "", nil
}

func (i *InstanceToken) WriteTokenFile(filePath, token string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Create(filePath); err != nil {
			return fmt.Errorf("failed to create token file: %v", err)
		}
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %v", filePath, err)
	}

	defer f.Close()

	if _, err = f.WriteString(token); err != nil {
		return fmt.Errorf("failed to write to file '%s': %v", filePath, err)
	}

	return nil
}

func (i *InstanceToken) WipeTokenFile(filePath string) error {
	if err := deleteFile(filePath); err != nil {
		return fmt.Errorf("error deleting token file '%s' to be wiped: %v", filePath, err)
	}

	if err := createFile(filePath); err != nil {
		return fmt.Errorf("error creating token file '%s' that was wiped: %v", filePath, err)
	}

	return nil
}

func deleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func createFile(path string) error {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		file, err := os.Create(path)

		if err != nil {
			return err
		}
		defer file.Close()
	}

	//Set permissions
	if err := os.Chmod(path, os.FileMode(0600)); err != nil {
		return fmt.Errorf("error changing permissons of file '%s' to 0600: %v", path, err)
	}

	return nil
}

func (i *InstanceToken) initTokenNew() error {
	exists, err := i.fileExists(i.InitTokenFilePath())
	if err != nil {
		return fmt.Errorf("error checking file exists: %v", err)
	}
	if !exists {
		return fmt.Errorf("no init token file: '%s' exiting.", i.InitTokenFilePath())
	}
	i.Log.Debugf("File exists: %s", i.InitTokenFilePath())
	initToken, err := i.TokenFromFile(i.InitTokenFilePath())
	if err != nil {
		return fmt.Errorf("error reading init token from file: %v", err)
	}
	if initToken == "" {
		return fmt.Errorf("init token was not read from file '%s' exiting", i.InitTokenFilePath())
	}

	i.Log.Debugf("init token found '%s' at '%s'", initToken, i.InitTokenFilePath())
	i.vaultClient.SetToken(initToken)

	policies, err := i.TokenPolicies()
	if err != nil {
		return fmt.Errorf("failed to find init token policies: %v", err)
	}

	newToken, err := i.createToken(policies)
	if err != nil {
		return err
	}
	i.SetToken(newToken)

	i.Log.Infof("New token: %s", i.Token())

	return nil
}

func (i *InstanceToken) TokenPolicies() (policies []string, err error) {
	s, err := i.TokenLookup()
	if err != nil {
		return nil, err
	}

	if s == nil {
		return nil, fmt.Errorf("no secret from init token lookup")
	}

	dat, ok := s.Data["policies"]
	if !ok {
		return nil, errors.New("failed to get policy data from init token lookup")
	}

	d, ok := dat.([]interface{})
	if !ok {
		return nil, errors.New("failed to convert data to []interface")
	}

	policies = make([]string, len(d))

	for n, m := range d {
		str, ok := m.(string)
		if !ok {
			return nil, errors.New("failed to convert interface to string")
		}
		policies[n] = str
	}

	return policies, nil
}

func (i *InstanceToken) createToken(policies []string) (token string, err error) {
	tCreateRequest := &vault.TokenCreateRequest{
		DisplayName: i.InitRole(),
	}

	newToken, err := i.vaultClient.Auth().Token().CreateWithRole(tCreateRequest, i.InitRole())
	if err != nil {
		return "", fmt.Errorf("failed to create init token: %v", err)
	}

	return newToken.Auth.ClientToken, nil
}

func (i *InstanceToken) TokenLookup() (secret *vault.Secret, err error) {

	s, err := i.vaultClient.Auth().Token().LookupSelf()
	if err != nil {
		return nil, fmt.Errorf("error lookup self token: %v", err)
	}

	if s == nil {
		return nil, errors.New("failed to find secret form Lookup self")
	}

	return s, nil
}

func (i *InstanceToken) tokenRenew() error {
	// Check if renewable

	s, err := i.TokenLookup()
	if err != nil {
		return err
	}

	dat, ok := s.Data["renewable"]
	if !ok {
		return errors.New("unable to get renewable token data from secret")
	}

	if dat == false {
		i.Log.Infof("Token not renewable: %s", i.Token())
		return nil
	}
	i.Log.Debugf("Token renewable")

	// Renew against vault
	s, err = i.vaultClient.Auth().Token().RenewSelf(0)
	if err != nil {
		return fmt.Errorf("error renewing token %s: %v", i.InitRole(), err)
	}

	i.Log.Infof("Renewed token: %s", i.Token())

	return nil
}

func (i *InstanceToken) EnsureToken() (newCreated bool, err error) {
	token, err := i.TokenRetrieve()
	if err != nil && os.IsExist(err) {
		return false, fmt.Errorf("error retrieving token from file: %v", err)
	}
	if token != "" {
		// Token exists in file
		logrus.Debugf("Token to renew: %s", token)
		i.SetToken(token)
		i.vaultClient.SetToken(i.Token())
		return false, nil
	}

	//Token Doesn't exist
	i.Log.Info("Token doesn't exist, generating new")
	err = i.initTokenNew()
	if err != nil {
		return false, fmt.Errorf("failed to generate new token: %v", err)
	}

	if err := i.WriteTokenFile(i.TokenFilePath(), i.Token()); err != nil {
		return false, fmt.Errorf("failed to write token to file: %v", err)
	}
	if err := i.WipeTokenFile(i.InitTokenFilePath()); err != nil {
		return false, fmt.Errorf("failed to wipe token from file: %v", err)
	}

	i.Log.Infof("Token written to file: %s", i.TokenFilePath())
	i.vaultClient.SetToken(i.Token())

	return true, nil
}

func (i *InstanceToken) TokenRenewRun() error {
	newCreated, err := i.EnsureToken()
	if err != nil {
		return err
	}

	if !newCreated {
		if err := i.tokenRenew(); err != nil {
			return err
		}
	}

	return nil
}
