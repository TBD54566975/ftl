package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"
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
		Engine:                   ptr("aurora-postgresql"),
		MasterUsername:           ptr("root"),
		ManageMasterUserPassword: ptr(true),
		DBSubnetGroupName:        ptr(p.config.DatabaseSubnetGroupARN),
		VpcSecurityGroupIds:      []string{p.config.DatabaseSecurityGroup},
		EngineMode:               ptr("provisioned"),
		Port:                     ptr(5432),
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
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "ReadEndpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyPsqlReadEndpoint,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "MasterUserSecret.SecretArn"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyPsqlMasterUserARN,
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

			adminEndpoint := endpointToDSN(write.OutputValue, "postgres", 5432, username, password)

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
			if _, err := db.ExecContext(ctx, "CREATE USER ftluser WITH LOGIN; GRANT rds_iam TO ftluser;"); err != nil {
				// Ignore if user already exists
				if !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("failed to create database: %w", err)
				}
			}
			if _, err := db.ExecContext(ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %s TO ftluser;", resourceID)); err != nil {
				return fmt.Errorf("failed to grant FTL user privileges: %w", err)
			}
		}
	}
	return nil
}

func updatePostgresOutputs(_ context.Context, to *schemapb.DatabaseRuntime, resourceID string, outputs []types.Output) error {
	byName, err := outputsByPropertyName(outputs)
	if err != nil {
		return fmt.Errorf("failed to group outputs by property name: %w", err)
	}

	to.WriteConnector = &schemapb.DatabaseConnector{
		Value: &schemapb.DatabaseConnector_AwsiamAuthDatabaseConnector{
			AwsiamAuthDatabaseConnector: &schemapb.AWSIAMAuthDatabaseConnector{
				Endpoint: *byName[PropertyPsqlWriteEndpoint].OutputValue,
				Database: resourceID,
				Username: "ftluser",
			},
		},
	}
	to.ReadConnector = &schemapb.DatabaseConnector{
		Value: &schemapb.DatabaseConnector_AwsiamAuthDatabaseConnector{
			AwsiamAuthDatabaseConnector: &schemapb.AWSIAMAuthDatabaseConnector{
				Endpoint: *byName[PropertyPsqlReadEndpoint].OutputValue,
				Database: resourceID,
				Username: "ftluser",
			},
		},
	}

	return nil
}
