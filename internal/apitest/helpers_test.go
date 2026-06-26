package apitest

import "go.uber.org/zap"

func zapNop() *zap.Logger { return zap.NewNop() }
