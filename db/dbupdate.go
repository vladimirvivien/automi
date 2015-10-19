package db

import (
	"database/sql"
	"fmt"

	"github.com/vladimirvivien/automi/api"
)

// DbUpdate is a processor that can submit SQL updates for items in its input.
// The compoenent allows for preparation of statement and handling of result.
// TODO - Rewrite how DbUpdate should work.
type DbUpdate struct {
	Name    string
	Input   <-chan interface{}
	DB      *sql.DB
	SQL     string
	Prepare func(interface{}) []interface{}
	Handle  func(interface{}, sql.Result) interface{}

	logs chan interface{}
	done chan struct{}
}

func (t *DbUpdate) Init() error {
	if t.Name == "" {
		return api.ProcError{Err: fmt.Errorf("Missing Name attribute")}
	}
	if t.Input == nil {
		return api.ProcError{
			ProcName: t.Name,
			Err:      fmt.Errorf("Missing Input attribute"),
		}
	}
	if t.DB == nil {
		return api.ProcError{
			ProcName: t.Name,
			Err:      fmt.Errorf("Missing DB attribute"),
		}
	}
	if t.SQL == "" {
		return api.ProcError{
			ProcName: t.Name,
			Err:      fmt.Errorf("Missing Input attribute"),
		}
	}

	if t.Prepare == nil {
		return api.ProcError{
			ProcName: t.Name,
			Err:      fmt.Errorf("Missing Prepare function"),
		}
	}

	t.logs = make(chan interface{})
	t.done = make(chan struct{})
	return nil
}

func (t *DbUpdate) Uninit() error {
	return nil
}

func (t *DbUpdate) GetName() string {
	return t.Name
}

func (t *DbUpdate) GetInput() <-chan interface{} {
	return t.Input
}

func (t *DbUpdate) Done() <-chan struct{} {
	return t.done
}

func (t *DbUpdate) GetLogs() <-chan interface{} {
	return t.logs
}

func (t *DbUpdate) Exec() error {
	go func() {
		defer func() {
			close(t.done)
			close(t.logs)
		}()
		stmt, err := t.DB.Prepare(t.SQL)
		if err != nil {
			panic(fmt.Sprintf("[%s] Prepare statement error: %s", t.Name, err))
		}

		for item := range t.Input {
			args := t.Prepare(item)

			tx, err := t.DB.Begin()
			if err != nil {
				panic(fmt.Sprintf("[%s] Transaction error: %s", t.Name, err))
			}

			// Submit transaction
			result, err := tx.Stmt(stmt).Exec(args)
			if err != nil {
				t.logs <- api.ProcError{ProcName: t.Name, Err: err}
				if rberr := tx.Rollback(); rberr != nil {
					t.logs <- api.ProcError{
						ProcName: t.Name,
						Err:      fmt.Errorf("Rollback failed: %s", rberr),
					}
				}
				continue
			}

			// handle result
			if t.Handle != nil {
				handle := t.Handle(item, result)
				switch h := handle.(type) {
				case nil:
					continue
				case api.ProcError, error:
					t.logs <- h
				}
			}

			if err := tx.Commit(); err != nil {
				t.logs <- api.ProcError{
					ProcName: t.Name,
					Err:      fmt.Errorf("Failed to commit: %s", err),
				}
				if rberr := tx.Rollback(); rberr != nil {
					t.logs <- api.ProcError{
						ProcName: t.Name,
						Err:      fmt.Errorf("Rollback failed: %s", rberr),
					}
				}
			}
		}
	}()

	return nil
}