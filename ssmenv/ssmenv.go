package ssmenv

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/vrischmann/envconfig"
)

type ssmConfig struct {
	Path     string `envconfig:"default=NOT_SET,SSM_PATH"`
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

	if path == "NOT_SET" {
		return fmt.Errorf("missing SSM_PATH environment variable")
	}

	if path == "" {
		return fmt.Errorf("wrong path configuration")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := ssm.New(sess)

	return setEnvVars(path, client)
}

func retryGetParameters(client *ssm.SSM, input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	count := 0
	for {
		output, err := client.GetParametersByPath(input)

		if err != nil {
			if count >= 5 {
				return nil, err
			}

			time.Sleep(5 * time.Second)
			count++
			continue
		}

		return output, nil
	}
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

		output, err := retryGetParameters(client, input)
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
