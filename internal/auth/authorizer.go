package auth

import (
	"fmt"

	"github.com/casbin/casbin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/comfforts/errors"
	"github.com/comfforts/logger"

	"github.com/comfforts/comff-stores/pkg/constants"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

type Authorizer struct {
	enforcer *casbin.Enforcer
	logger   logger.AppLogger
}

func NewAuthorizer(model, policy string, logger logger.AppLogger) (*Authorizer, error) {
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
		a.logger.Error(msg, zap.Error(constants.ErrUserAccessDenied))
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	return nil
}
