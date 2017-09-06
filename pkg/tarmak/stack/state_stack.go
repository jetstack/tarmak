package stack

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type StateStack struct {
	*Stack
}

var _ interfaces.Stack = &StateStack{}

func newStateStack(s *Stack) (*StateStack, error) {
	ss := &StateStack{
		Stack: s,
	}

	s.name = tarmakv1alpha1.StackNameState
	s.verifyPostDeploy = append(s.verifyPostDeploy, ss.verifyDNSDelegation)
	return ss, nil
}

func (s *StateStack) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	// TODO: refactor me
	/*
		state := s.Stack.conf.State
		if state.BucketPrefix != "" {
			output["bucket_prefix"] = state.BucketPrefix
		}
		if state.PublicZone != "" {
			output["public_zone"] = state.PublicZone
		}
	*/

	return output
}

func (s *StateStack) VerifyPost() error {
	return s.verifyDNSDelegation()
}

func (s *StateStack) verifyDNSDelegation() error {

	tries := 5
	for {
		// TODO: refactor me
		//host := strings.Join([]string{utils.RandStringRunes(16), "_tarmak", s.conf.State.PublicZone}, ".")
		host := "refactor.me"

		result, err := net.LookupTXT(host)
		if err == nil {
			if reflect.DeepEqual([]string{"tarmak delegation works"}, result) {
				return nil
			} else {
				s.log.Warn("error checking delegation to public zone: ", err)
			}
		} else {
			s.log.Warn("error checking delegation to public zone: ", err)
		}

		if tries == 0 {
			nameservers, ok := s.Output()["public_zone_name_servers"]
			msg := "failed verifying delegation of public zone 5 times"
			if ok {
				// TODO: refactor me
				msg = fmt.Sprintf("%s, make sure the zone %s is delegated to nameservers %s", msg, host, nameservers)
				//msg = fmt.Sprintf("%s, make sure the zone %s is delegated to nameservers %s", msg, s.conf.State.PublicZone, nameservers)
			}

			return errors.New(msg)
		}
		tries -= 1
		time.Sleep(2 * time.Second)
	}
}
