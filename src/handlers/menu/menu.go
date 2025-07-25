package menu

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
