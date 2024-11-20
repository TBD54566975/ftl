package main

import (
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"
)

type MySQLTemplater struct {
	resourceID string
	cluster    string
	module     string
	config     *Config
}

var _ ResourceTemplater = (*MySQLTemplater)(nil)

func (p *MySQLTemplater) AddToTemplate(template *goformation.Template) error {
	clusterID := cloudformationResourceID(p.resourceID, "cluster")
	instanceID := cloudformationResourceID(p.resourceID, "instance")
	template.Resources[clusterID] = &rds.DBCluster{
		Engine:                   ptr("aurora-mysql"),
		MasterUsername:           ptr("root"),
		ManageMasterUserPassword: ptr(true),
		DBSubnetGroupName:        ptr(p.config.DatabaseSubnetGroupARN),
		VpcSecurityGroupIds:      []string{p.config.DatabaseSecurityGroup},
		EngineMode:               ptr("provisioned"),
		Port:                     ptr(3306),
		ServerlessV2ScalingConfiguration: &rds.DBCluster_ServerlessV2ScalingConfiguration{
			MinCapacity: ptr(0.5),
			MaxCapacity: ptr(10.0),
		},
		Tags: ftlTags(p.cluster, p.module),
	}
	template.Resources[instanceID] = &rds.DBInstance{
		Engine:              ptr("aurora-mysql"),
		DBInstanceClass:     ptr("db.serverless"),
		DBClusterIdentifier: ptr(goformation.Ref(clusterID)),
		Tags:                ftlTags(p.cluster, p.module),
	}
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "Endpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyMySQLWriteEndpoint,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "ReadEndpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyMySQLReadEndpoint,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "MasterUserSecret.SecretArn"), &CloudformationOutputKey{
		ResourceID:   p.resourceID,
		PropertyName: PropertyMySQLMasterUserARN,
	})
	return nil
}
