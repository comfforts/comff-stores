package auth

import (
	"fmt"

	"github.com/casbin/casbin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

type Authorizer struct {
	enforcer *casbin.Enforcer
	logger   *logging.AppLogger
}

func NewAuthorizer(model, policy string, logger *logging.AppLogger) (*Authorizer, error) {
	_, err := fileUtils.FileStats(model)
	if err != nil {
		msg := fmt.Sprintf(fileUtils.ERROR_NO_FILE, model)
		logger.Error(msg, zap.Error(err))
		return nil, errors.WrapError(err, fileUtils.ERROR_NO_FILE, model)
	}

	_, err = fileUtils.FileStats(policy)
	if err != nil {
		msg := fmt.Sprintf(fileUtils.ERROR_NO_FILE, model)
		logger.Error(msg, zap.Error(err))
		return nil, errors.WrapError(err, fileUtils.ERROR_NO_FILE, policy)
	}

	enforcer := casbin.NewEnforcer(model, policy)
	return &Authorizer{
		enforcer: enforcer,
		logger:   logger,
	}, nil
}

func (a *Authorizer) Authorize(subject, object, action string) error {
	if !a.enforcer.Enforce(subject, object, action) {
		msg := fmt.Sprintf("%s not permitted to %s to %s", subject, action, object)
		a.logger.Error(msg, zap.Error(errors.ErrUserAccessDenied))
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	return nil
}
