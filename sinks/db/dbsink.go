package db

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
)

type DbSink struct {
	input  <-chan interface{}
	db     api.Database
	log    *logrus.Entry
	sql    string
	prepFn func(interface{}) []interface{}
}

func New() *DbSink {
	return new(DbSink)
}

func (s *DbSink) WithDB(db api.Database) *DbSink {
	s.db = db
	return s
}

func (s *DbSink) Prepare(fn func(interface{}) []interface{}) *DbSink {
	s.prepFn = fn
	return s
}

func (s *DbSink) Sql(sql string) *DbSink {
	s.sql = sql
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

	if s.sql == "" {
		return fmt.Errorf("Sql attribute is required")
	}

	if s.prepFn == nil {
		return fmt.Errorf("Prepare function is missing")
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

	go func() {
		defer func() {
			close(result)
			s.db.Close()
			s.log.Info("Component closed")
		}()
		stmt, err := s.db.Prepare(s.sql)
		if err != nil {
			result <- err
			return
		}

		for item := range s.input {
			args := s.prepFn(item) // prepare sql args

			tx, err := s.db.Begin()
			if err != nil {
				result <- err
				return
			}

			// exec sql within tx
			_, err = tx.Stmt(stmt).Exec(args)
			if err != nil {
				s.log.Error(err) // log sql error, continue
				if rberr := tx.Rollback(); rberr != nil {
					// something maybe wrong, stop
					s.log.Error(rberr)
					result <- rberr
					return
				}
				continue
			}

			// commit tx
			if err := tx.Commit(); err != nil {
				s.log.Error(err)
				if rberr := tx.Rollback(); rberr != nil {
					s.log.Errorf("Rollback failed %v", err)
					result <- err
					return
				}
				continue
			}
		}
	}()
	return result
}
