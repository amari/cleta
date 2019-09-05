/*
Copyright © 2019 Amari Robinson

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	// logging
	log *zap.Logger
	// tracing
}

func NewServer(core *zapcore.Core) (*Server, error) {
	return nil, nil
}

func NewNopServer() (*Server, error) {
	return nil, nil
}

func NewDevelopmentServer(options ...zap.Option) (*Server, error) {
	l, err := zap.NewDevelopment(options...)
	if err != nil {
		return nil, err
	}
	return &Server{
		log: l,
	}, nil
}

func NewProductionServer() (*Server, error) {
	return nil, nil
}

func (s *Server) WithLoggerFields(fields ...zap.Field) *Server {
	return &Server{
		log: s.Log().With(fields...),
	}
}

func (s *Server) WithLoggerOptions(opts ...zap.Option) *Server {
	return &Server{
		log: s.Log().WithOptions(opts...),
	}
}

func (s *Server) Log() *zap.Logger {
	return s.log
}
