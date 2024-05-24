package configuration

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"log"
	"net/url"
)

type AWSSecrets[R Role] struct {
	client *secretsmanager.Client

	AccessKeyId     string
	SecretAccessKey string
	Region          string
	Endpoint        string
}

var _ Resolver[Secrets] = AWSSecrets[Secrets]{}
var _ Provider[Secrets] = AWSSecrets[Secrets]{}
var _ MutableProvider[Secrets] = AWSSecrets[Secrets]{}

func (a AWSSecrets[R]) getClient(ctx context.Context) (*AWSSecrets[Secrets], error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(a.region),
	)
	if err != nil {
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		e, ok := endpoint.Get()
		if ok {
			o.BaseEndpoint = aws.String(e)
		}
	})

	return &AWSSecrets[Secrets]{
		AccessKeyId:     accessKeyId,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		client:          svc,
	}, nil
}

func (a AWSSecrets[R]) Role() R {
	var r R
	return r
}

func (a AWSSecrets[R]) Get(ctx context.Context, ref Ref) (*url.URL, error) {}

func (a AWSSecrets[R]) Set(ctx context.Context, ref Ref, key *url.URL) error {
	//this will have to do a list/check to see if the secret exists

}

func (a AWSSecrets[R]) Unset(ctx context.Context, ref Ref) error {}

func (a AWSSecrets[R]) List(ctx context.Context) ([]Entry, error) {}

func (a AWSSecrets[R]) Key() string {
	return "asm"
}

func (a AWSSecrets[R]) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {}

func (a AWSSecrets[R]) Writer() bool {
	return true
}

func (a AWSSecrets[R]) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {}

func (a AWSSecrets[R]) Delete(ctx context.Context, ref Ref) error {}

func brainstorm() {
	endpoint := "http://localhost:4566"

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
	)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	// create a secret
	name := "test-secret3"
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
