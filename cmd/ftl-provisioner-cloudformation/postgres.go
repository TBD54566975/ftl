package main

import (
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
		PropertyName: PropertyDBWriteEndpoint,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "ReadEndpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyDBReadEndpoint,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "MasterUserSecret.SecretArn"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyMasterUserARN,
	})
	return nil
}
