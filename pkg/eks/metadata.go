package eks

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

// EC2MetadataClient interface
type EC2MetadataClient struct {
	*ec2metadata.EC2Metadata
}

// NewEC2MetadataClient instantiates an EC2 Metadata client
func NewEC2MetadataClient() (*EC2MetadataClient, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	md := ec2metadata.New(sess)
	return &EC2MetadataClient{md}, nil
}

// GetUserData wraps the the AWS EC2 Metadata call
// This is all so we can mock it. There has to be a better way but
// the AWS Go SDK seems...a bit rubbish.
func (c *EC2MetadataClient) GetUserData() (string, error) {
	return c.GetUserData()
}

// Region wraps the the AWS EC2 Region call
// This is all so we can mock it. There has to be a better way but
// the AWS Go SDK seems...a bit rubbish.
func (c *EC2MetadataClient) Region() (string, error) {
	return c.Region()
}
