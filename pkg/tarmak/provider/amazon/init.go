// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

func Init(in *input.Input, provider *tarmakv1alpha1.Provider) error {
	if provider.Amazon == nil {
		provider.Amazon = &tarmakv1alpha1.ProviderAmazon{}
	}
	err := initCredentials(in, provider)
	if err != nil {
		return err
	}

	err = initBucketPrefix(in, provider)
	if err != nil {
		return err
	}

	err = initPublicZone(in, provider)
	if err != nil {
		return err
	}

	return nil
}

func initBucketPrefix(in *input.Input, provider *tarmakv1alpha1.Provider) error {
	for {
		bucketPrefix, err := in.AskOpen(&input.AskOpen{
			Query:   "Which prefix should be used for the state buckets and DynamoDB tables? ([a-z0-9-]+, should be globally unique)",
			Default: fmt.Sprintf("%s-tarmak-", provider.Name),
		})
		if err != nil {
			return err
		}

		nameValid := input.RegexpName.MatchString(bucketPrefix)

		if !nameValid {
			in.Warnf("bucket/table prefix '%s' is not valid", bucketPrefix)
		} else {
			provider.Amazon.BucketPrefix = bucketPrefix
			break
		}
	}

	return nil
}

func initPublicZone(in *input.Input, provider *tarmakv1alpha1.Provider) error {
	for {
		publicZone, err := in.AskOpen(&input.AskOpen{
			Query: "Which public DNS zone should be used? (the DNS zone will be created if it does not exist and it must be delegated from the root)",
		})
		if err != nil {
			return err
		}

		zoneValid := input.RegexpDNS.MatchString(publicZone)

		if !zoneValid {
			in.Warnf("Public DNS zone '%s' is not valid", publicZone)
		} else {
			provider.Amazon.PublicZone = publicZone
			break
		}
	}

	return nil
}

func initCredentials(in *input.Input, provider *tarmakv1alpha1.Provider) error {

	credentialSources := []string{
		"Amazon CLI auth, using environment variables or profiles in '~/.aws'",
		"read from Vault path",
	}

	credentialSource, err := in.AskSelection(&input.AskSelection{
		Query:   "Where should the credentials for this provider come from?",
		Choices: credentialSources,
		Default: 0,
	})
	if err != nil {
		return err
	}

	// AWS folder
	if credentialSource == 0 {
		for {
			awsProfile, err := in.AskOpen(&input.AskOpen{
				Query:      "Which AWS profile should be used? (leave empty for default profile)",
				AllowEmpty: true,
			})
			if err != nil {
				return err
			}
			provider.Amazon.Profile = awsProfile
			break
		}
	}

	// Vault Path
	if credentialSource == 1 {
		for {
			vaultPath, err := in.AskOpen(&input.AskOpen{
				Query: "Which Vault path should be used for Amazon credentials?",
			})
			if err != nil {
				return err
			}
			provider.Amazon.VaultPath = vaultPath
			break
		}
	}

	return nil

}
