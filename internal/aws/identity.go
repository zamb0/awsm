package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// CallerIdentity is the result of an sts:GetCallerIdentity call.
type CallerIdentity struct {
	Account string
	Arn     string
	UserID  string
}

// GetCallerIdentity calls sts:GetCallerIdentity using either the provided
// profile (loaded via the shared AWS config) or the provided temporary
// credentials. If both are empty/nil, it uses the default credential chain.
//
// When creds is provided, it takes precedence and the call uses those exact
// credentials. The region is used to construct the STS client when creds is
// supplied; if empty, "us-east-1" is used as a safe default for STS (which is
// a global service but the SDK still requires a region).
func GetCallerIdentity(ctx context.Context, profile string, creds *TempCredentials, region string) (*CallerIdentity, error) {
	var (
		cfg aws.Config
		err error
	)

	if creds != nil {
		opts := []func(*config.LoadOptions) error{
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken,
			)),
		}
		if region != "" {
			opts = append(opts, config.WithRegion(region))
		} else {
			opts = append(opts, config.WithRegion("us-east-1"))
		}
		cfg, err = config.LoadDefaultConfig(ctx, opts...)
	} else if profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sts.NewFromConfig(cfg)
	out, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		if isSsoExpiredErr(err) {
			return nil, ErrSsoSessionExpired
		}
		return nil, fmt.Errorf("sts:GetCallerIdentity failed: %w", err)
	}

	id := &CallerIdentity{}
	if out.Account != nil {
		id.Account = *out.Account
	}
	if out.Arn != nil {
		id.Arn = *out.Arn
	}
	if out.UserId != nil {
		id.UserID = *out.UserId
	}
	return id, nil
}

func isSsoExpiredErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrSsoSessionExpired) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "token has expired") ||
		strings.Contains(msg, "InvalidGrantException") ||
		strings.Contains(msg, "expired")
}
