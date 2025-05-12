package authManager

import (
	"sync"

	auth_v1 "github.com/balobas/sport_city_common/api/auth_service_v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientsAuthManager struct {
	cfg    Config
	cc     *grpc.ClientConn
	client auth_v1.AuthClient

	mu         *sync.RWMutex
	accessJwt  string
	refreshJwt string
}

type Config interface {
	ServiceCreds() (email string, pwd string)
	AuthServiceGrpcAddress() string
}

func New(
	cfg Config,
) (*ClientsAuthManager, error) {
	cc, err := grpc.NewClient(cfg.AuthServiceGrpcAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &ClientsAuthManager{
		cfg:    cfg,
		cc:     cc,
		client: auth_v1.NewAuthClient(cc),
		mu:     &sync.RWMutex{},
	}, nil
}

func (cm *ClientsAuthManager) Close() error {
	return cm.cc.Close()
}
