package read

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/vault-helper/pkg/instanceToken"
)

const FlagOutputPath = "dest-path"
const FlagField = "field"
const FlagOwner = "owner"
const FlagGroup = "group"

type Read struct {
	vaultPath string
	fieldName string
	filePath  string
	owner     string
	group     string

	Log           *logrus.Entry
	instanceToken *instanceToken.InstanceToken
}

func (r *Read) RunRead() error {
	//Read vault
	sec, err := r.InstanceToken().VaultClient().Logical().Read(r.VaultPath())
	if err != nil {
		return fmt.Errorf("error reading from vault: %v", err)
	}

	if sec == nil {
		return errors.New("vault returned nothing")
	}

	var res string
	//Just get field
	if r.FieldName() != "" {
		res, err = r.getField(sec)
	} else {
		res, err = r.getPrettyJSON(sec)
	}
	if err != nil {
		return err
	}

	//Output to console
	if r.FilePath() == "" {
		str := ""
		if r.FieldName() != "" {
			str = "(" + r.FieldName() + ")"
		}
		str = "No file given. Outputting to console. " + str
		r.Log.Info(str)

		r.Log.Info(res)

		return nil
	}

	//Write to file
	r.Log.Infof("Outputing responce to file: %s", r.filePath)
	return r.writeToFile(res)
}

func (r *Read) getField(sec *vault.Secret) (field string, err error) {
	dat := sec.Data

	fieldDat, ok := dat[r.FieldName()]
	if !ok {
		return "", errors.New("error extracting field data from responce")
	}

	field, ok = fieldDat.(string)
	if !ok {
		b, ok := fieldDat.(bool)
		if !ok {
			i, ok := fieldDat.(json.Number)
			if !ok {
				return "", fmt.Errorf("error converting field data into string: %s", r.FieldName())
			}
			return string(i), nil
		}
		return strconv.FormatBool(b), nil
	}

	return field, nil
}

func (r *Read) writeToFile(res string) error {

	byt := []byte(res)
	if err := ioutil.WriteFile(r.FilePath(), byt, 0600); err != nil {
		return fmt.Errorf("error trying to write responce to file '%s': %s", r.FilePath(), err)
	}

	return r.writePermissons()
}

func (r *Read) writePermissons() error {

	if err := os.Chmod(r.FilePath(), os.FileMode(0600)); err != nil {
		return fmt.Errorf("error changing permissons of file '%s' to 0600: %v", r.FilePath(), err)
	}

	var uid int
	var gid int
	var err error
	var curr *user.User

	if r.Owner() == "" {
		r.Log.Debugf("No owner given. Defaulting permissions to current user")
		if curr, err = user.Current(); err != nil {
			return fmt.Errorf("error retrieving current user info: %v", err)
		}

		if uid, err = strconv.Atoi(curr.Uid); err != nil {
			return fmt.Errorf("failed to convert user uid '%s' (string) to (int): %v", curr.Uid, err)
		}

	} else if u, err := strconv.Atoi(r.Owner()); err == nil {
		r.Log.Debugf("User is a number. Using instead of lookup user")
		uid = u

	} else {
		usr, err := user.Lookup(r.Owner())
		if err != nil {
			return fmt.Errorf("failed to find user '%s' on system: %v", r.Owner(), err)
		}

		if uid, err = strconv.Atoi(usr.Uid); err != nil {
			return fmt.Errorf("failed to convert user uid '%s' (string) to (int): %v", usr.Uid, err)
		}
	}

	if r.Group() == "" {
		r.Log.Debugf("No group given. Defaulting permissions to current user-group")
		if curr == nil {
			if curr, err = user.Current(); err != nil {
				return fmt.Errorf("error retrieving current user info: %v", err)
			}
		}

		if gid, err = strconv.Atoi(curr.Gid); err != nil {
			return fmt.Errorf("failed to convert user gid '%s' (string) to (int): %v", curr.Gid, err)
		}

	} else if g, err := strconv.Atoi(r.Group()); err == nil {
		r.Log.Debugf("Group is a number. Using as gid instead of lookup group")
		gid = g

	} else {
		grp, err := user.LookupGroup(r.Group())
		if err != nil {
			return fmt.Errorf("failed to find group '%s' on system: %v", r.Group(), err)
		}

		if gid, err = strconv.Atoi(grp.Gid); err != nil {
			return fmt.Errorf("failed to convert group gid '%s' (string) to (int): %v", grp.Gid, err)
		}
	}

	if err := os.Chown(r.FilePath(), uid, gid); err != nil {
		return fmt.Errorf("failed to change group and owner of file '%s' to usr:'%s' grp:'%s': %v", r.FilePath(), r.Owner(), r.Group(), err)
	}

	r.Log.Debugf("Set permissons on file: %s", r.FilePath())

	return nil
}

func (r *Read) getPrettyJSON(sec *vault.Secret) (prettyStr string, err error) {
	js, err := json.Marshal(sec)
	if err != nil {
		return "", fmt.Errorf("error converting responce from vault into JSON: %v", err)
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, js, "", "\t")
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	return string(prettyJSON.Bytes()), nil
}

func New(log *logrus.Entry, i *instanceToken.InstanceToken) *Read {
	r := &Read{
		instanceToken: i,
		Log:           log,
	}

	if log != nil {
		r.Log = log
	}

	return r
}

func (r *Read) SetVaultPath(path string) {
	r.vaultPath = path
}
func (r *Read) VaultPath() (path string) {
	return r.vaultPath
}

func (r *Read) SetFieldName(name string) {
	r.fieldName = name
}
func (r *Read) FieldName() (name string) {
	return r.fieldName
}

func (r *Read) SetFilePath(path string) {
	r.filePath = path
}
func (r *Read) FilePath() (path string) {
	return r.filePath
}

func (r *Read) SetOwner(name string) {
	r.owner = name
}
func (r *Read) Owner() (name string) {
	return r.owner
}

func (r *Read) SetGroup(name string) {
	r.group = name
}
func (r *Read) Group() (name string) {
	return r.group
}

func (r *Read) InstanceToken() *instanceToken.InstanceToken {
	return r.instanceToken
}
