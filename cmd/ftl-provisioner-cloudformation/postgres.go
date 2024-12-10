package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"

	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	"github.com/TBD54566975/ftl/internal/schema"
)

type PostgresTemplater struct {
	resourceID string
	cluster    string
	module     string
	config     *Config
}

var _ ResourceTemplater = (*PostgresTemplater)(nil)

func (p *PostgresTemplater) AddToTemplate(template *goformation.Template) error {
	clusterID := cloudformationResourceID(p.resourceID, "cluster")
	instanceID := cloudformationResourceID(p.resourceID, "instance")
	template.Resources[clusterID] = &rds.DBCluster{
		Engine:                          ptr("aurora-postgresql"),
		MasterUsername:                  ptr("root"),
		ManageMasterUserPassword:        ptr(true),
		DBSubnetGroupName:               ptr(p.config.DatabaseSubnetGroupARN),
		VpcSecurityGroupIds:             []string{p.config.DatabaseSecurityGroup},
		EngineMode:                      ptr("provisioned"),
		Port:                            ptr(5432),
		EnableIAMDatabaseAuthentication: ptr(true),
		ServerlessV2ScalingConfiguration: &rds.DBCluster_ServerlessV2ScalingConfiguration{
			MinCapacity: ptr(0.5),
			MaxCapacity: ptr(10.0),
		},
		Tags: ftlTags(p.cluster, p.module),
	}
	template.Resources[instanceID] = &rds.DBInstance{
		Engine:              ptr("aurora-postgresql"),
		DBInstanceClass:     ptr("db.serverless"),
		DBClusterIdentifier: ptr(goformation.Ref(clusterID)),
		Tags:                ftlTags(p.cluster, p.module),
	}
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "Endpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyPsqlWriteEndpoint,
		ResourceKind: ResourceKindPostgres,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "ReadEndpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyPsqlReadEndpoint,
		ResourceKind: ResourceKindPostgres,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "MasterUserSecret.SecretArn"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyPsqlMasterUserARN,
		ResourceKind: ResourceKindPostgres,
	})
	return nil
}

func PostgresPostUpdate(ctx context.Context, secrets *secretsmanager.Client, byName map[string]types.Output, resourceID string) error {
	if write, ok := byName[PropertyPsqlWriteEndpoint]; ok {
		if secret, ok := byName[PropertyPsqlMasterUserARN]; ok {
			secretARN := *secret.OutputValue
			username, password, err := secretARNToUsernamePassword(ctx, secrets, secretARN)
			if err != nil {
				return fmt.Errorf("failed to get username and password from secret ARN: %w", err)
			}

			if err := createPostgresDatabase(ctx, *write.OutputValue, resourceID, username, password); err != nil {
				return fmt.Errorf("failed to create postgres database: %w", err)
			}
			adminEndpoint := endpointToDSN(write.OutputValue, resourceID, 5432, username, password)
			db, err := sql.Open("pgx", adminEndpoint)
			if err != nil {
				return fmt.Errorf("failed to connect to postgres: %w", err)
			}
			defer db.Close()

			// Create the database if it doesn't exist
			if _, err := db.ExecContext(ctx, "CREATE DATABASE "+resourceID); err != nil {
				// Ignore if database already exists
				if !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("failed to create database: %w", err)
				}
			}
			if _, err := db.ExecContext(ctx, "CREATE USER ftluser WITH LOGIN; GRANT rds_iam TO ftluser;"); err != nil {
				// Ignore if user already exists
				if !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("failed to create database: %w", err)
				}
			}
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				GRANT CONNECT ON DATABASE %s TO ftluser;
				GRANT USAGE ON SCHEMA public TO ftluser;
				GRANT CREATE ON SCHEMA public TO ftluser;
				GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ftluser;
				GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ftluser;
				ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ftluser;
				ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ftluser;
			`, resourceID)); err != nil {
				return fmt.Errorf("failed to grant FTL user privileges: %w", err)
			}
		}
	}
	return nil
}

func createPostgresDatabase(ctx context.Context, endpoint, resourceID, username, password string) error {
	adminEndpoint := endpointToDSN(&endpoint, "postgres", 5432, username, password)

	// Connect to postgres without a specific database to create the new one
	db, err := sql.Open("pgx", adminEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// Create the database if it doesn't exist
	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+resourceID); err != nil {
		// Ignore if database already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	return nil
}

func updatePostgresOutputs(_ context.Context, resourceID string, outputs []types.Output) ([]*provisioner.ProvisioningEvent, error) {
	byName, err := outputsByPropertyName(outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to group outputs by property name: %w", err)
	}

	event := schema.DatabaseRuntimeEvent{
		ID: resourceID,
		Payload: &schema.DatabaseRuntimeConnectionsEvent{
			Connections: &schema.DatabaseRuntimeConnections{
				Write: &schema.AWSIAMAuthDatabaseConnector{
					Endpoint: fmt.Sprintf("%s:%d", *byName[PropertyPsqlWriteEndpoint].OutputValue, 5432),
					Database: resourceID,
					Username: "ftluser",
				},
				Read: &schema.AWSIAMAuthDatabaseConnector{
					Endpoint: fmt.Sprintf("%s:%d", *byName[PropertyPsqlReadEndpoint].OutputValue, 5432),
					Database: resourceID,
					Username: "ftluser",
				},
			},
		},
	}
	return []*provisioner.ProvisioningEvent{{
		Value: &provisioner.ProvisioningEvent_DatabaseRuntimeEvent{
			DatabaseRuntimeEvent: event.ToProto(),
		},
	}}, nil
}
