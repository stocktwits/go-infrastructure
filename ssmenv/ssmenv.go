package ssmenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/vrischmann/envconfig"
)

type ssmConfig struct {
	Path     string `envconfig:"SSM_PATH"`
	Disabled bool   `envconfig:"default=False,SSM_DISABLED"`
}

//Loads the SSM singleton instance and calls MustProcess
func InitEnvVars() error {
	cfg := &ssmConfig{}
	err := envconfig.Init(cfg)
	if err != nil {
		return err
	}

	if cfg.Disabled {
		return nil
	}

	path := cfg.Path
	if path == "" {
		return fmt.Errorf("wrong path configuration")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := ssm.New(sess)

	return setEnvVars(path, client)
}

func setEnvVars(path string, client *ssm.SSM) error {

	var nextToken *string
	for {
		input := &ssm.GetParametersByPathInput{
			WithDecryption: aws.Bool(true),
			Recursive:      aws.Bool(true),
			Path:           aws.String(path),
			NextToken:      nextToken,
		}

		output, err := client.GetParametersByPath(input)
		if err != nil {
			err = fmt.Errorf("error connecting to ssm store %v", err)
			return err
		}

		for _, param := range output.Parameters {
			k := strings.Replace(*param.Name, path, "", 1)
			k = strings.ToUpper(k)
			v := *param.Value
			err := os.Setenv(k, v)
			if err != nil {
				errR := fmt.Errorf("problem copying ssm key (%s) to environment variable (%s) - %v", *param.Name, k, err)
				return errR
			}
		}
		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	return nil

}
