package awsutil

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	"github.com/rebuy-de/aws-nuke/pkg/config"
)

type Account struct {
	Credentials
	CustomEndpoints config.CustomEndpoints

	id      string
	aliases []string
}

func NewAccount(creds Credentials, endpoints config.CustomEndpoints) (*Account, error) {
	account := Account{
		Credentials:     creds,
		CustomEndpoints: endpoints,
	}

	defaultSession, err := account.NewSession(DefaultRegionID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create default session in %s", DefaultRegionID)
	}

	customStackSupportSTSAndIAM := true
	if endpoints.GetRegion(DefaultRegionID) != nil {
		if endpoints.GetURL(DefaultRegionID, "sts") == "" {
			customStackSupportSTSAndIAM = false
		} else if endpoints.GetURL(DefaultRegionID, "iam") == "" {
			customStackSupportSTSAndIAM = false
		}
	}
	if !customStackSupportSTSAndIAM {
		account.aliases = []string{"Your account for the custom region " + DefaultRegionID}
		account.id = "account-id-of-custom-region-" + DefaultRegionID
		return &account, nil
	}

	identityOutput, err := sts.New(defaultSession).GetCallerIdentity(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed get caller identity")
	}

	globalSession, err := account.NewSession(GlobalRegionID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create global session in %s", GlobalRegionID)
	}

	aliasesOutput, err := iam.New(globalSession).ListAccountAliases(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed get account alias")
	}

	aliases := []string{}
	for _, alias := range aliasesOutput.AccountAliases {
		if alias != nil {
			aliases = append(aliases, *alias)
		}
	}

	account.id = *identityOutput.Account
	account.aliases = aliases

	return &account, nil
}

func (a *Account) ID() string {
	return a.id
}

func (a *Account) Alias() string {
	return a.aliases[0]
}

func (a *Account) Aliases() []string {
	return a.aliases
}
