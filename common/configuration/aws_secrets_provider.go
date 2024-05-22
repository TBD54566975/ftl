package configuration

import (
	"context"
	_ "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"log"
)

func brainstorm() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := secretsmanager.NewFromConfig(cfg)

	// create a secret
	name := "test-secret"
	secret := "hunter1"
	_, err = svc.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
		Name:         &name,
		SecretString: &secret,
	})
	if err != nil {
		log.Fatalf("failed to create secret, %v", err)
	}

	// get the secret
	out, err := svc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &name,
	})
	if err != nil {
		log.Fatalf("failed to retrieve secret, %v", err)
	}

	log.Printf("retrieved secret: %s", *out.SecretString)

}
