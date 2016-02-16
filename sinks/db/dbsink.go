package db

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
)

type DbSink struct {
	input <-chan interface{}
	db    api.Database
	log   *logrus.Entry
}

func New() *DbSink {
	return new(DbSink)
}

func (s *DbSink) WithDB(db api.Database) *DbSink {
	s.db = db
	return s
}

func (s *DbSink) SetInput(in <-chan interface{}) {
	s.input = in
}

func (s *DbSink) init(ctx context.Context) error {
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "CsvSink")
		log.Error("No logger found in context")
	}

	s.log = log.WithFields(logrus.Fields{
		"Component": "DbSink",
		"Type":      fmt.Sprintf("%T", s),
	})

	if s.input == nil {
		return fmt.Errorf("Input attribute not set")
	}

	s.log.Info("Component initialized")

	return nil

}

func (s *DbSink) Open(ctx context.Context) <-chan error {
	result := make(chan error)
	if err := s.init(ctx); err != nil {
		go func() {
			result <- err
		}()
		return result
	}
	return result
}
