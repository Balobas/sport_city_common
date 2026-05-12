package authManager

import (
	"sync"

	"github.com/balobas/sport_city_common/api/auth_internal_api"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientsAuthManager struct {
	cfg    Config
	cc     *grpc.ClientConn
	client auth_internal_api.AuthInternalApiClient

	mu         *sync.RWMutex
	accessJwt  string
	refreshJwt string
}

type Config interface {
	ServiceCreds() (uid string, pwd string)
	ServiceName() string
	ServiceDomain() string
	AuthServiceGrpcAddress() string
	ServiceDeviceUid() uuid.UUID
	ServiceDeviceName() string
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
		client: auth_internal_api.NewAuthInternalApiClient(cc),
		mu:     &sync.RWMutex{},
	}, nil
}

func (cm *ClientsAuthManager) Close() error {
	return cm.cc.Close()
}
