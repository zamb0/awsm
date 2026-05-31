package aws

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// awsEnvKeys are the environment variables the AWS SDK / CLI consult for
// credentials and configuration. We unset these in the parent environment
// before re-applying our own values so we don't leak state from the caller.
var awsEnvKeys = []string{
	"AWS_PROFILE",
	"AWS_DEFAULT_PROFILE",
	"AWS_ACCESS_KEY_ID",
	"AWS_SECRET_ACCESS_KEY",
	"AWS_SESSION_TOKEN",
	"AWS_SECURITY_TOKEN",
	"AWS_REGION",
	"AWS_DEFAULT_REGION",
	"AWS_CREDENTIAL_EXPIRATION",
}

// BuildEnvForProfile returns a copy of the current process environment with
// AWS_* variables scrubbed and the provided credentials, region and profile
// name injected. The result is suitable for passing to exec.Cmd.Env.
//
// If creds is nil, only AWS_PROFILE/AWS_REGION are set (lets the child use the
// shared config file directly).
func BuildEnvForProfile(creds *TempCredentials, region, profile string) []string {
	out := make([]string, 0, len(os.Environ())+8)
	for _, kv := range os.Environ() {
		k := kv
		if idx := strings.IndexByte(kv, '='); idx >= 0 {
			k = kv[:idx]
		}
		if isAWSEnvKey(k) {
			continue
		}
		out = append(out, kv)
	}

	if profile != "" {
		out = append(out, "AWS_PROFILE="+profile)
	}
	if region != "" {
		out = append(out, "AWS_REGION="+region)
		out = append(out, "AWS_DEFAULT_REGION="+region)
	}
	if creds != nil {
		out = append(out, "AWS_ACCESS_KEY_ID="+creds.AccessKeyId)
		out = append(out, "AWS_SECRET_ACCESS_KEY="+creds.SecretAccessKey)
		if creds.SessionToken != "" {
			out = append(out, "AWS_SESSION_TOKEN="+creds.SessionToken)
		}
		if !creds.Expires.IsZero() {
			out = append(out, fmt.Sprintf("AWS_CREDENTIAL_EXPIRATION=%s",
				creds.Expires.UTC().Format(time.RFC3339)))
		}
	}
	return out
}

func isAWSEnvKey(k string) bool {
	for _, key := range awsEnvKeys {
		if k == key {
			return true
		}
	}
	return false
}
