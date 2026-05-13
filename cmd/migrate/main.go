package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	if len(os.Args) < 2 {
		fail(errors.New("usage: migrate [up|down|force <v>]"))
	}

	m, err := migrate.New("file://migrations", cfg.DB.URL)
	if err != nil {
		fail(err)
	}
	defer m.Close()

	switch os.Args[1] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "force":
		if len(os.Args) < 3 {
			fail(errors.New("force requires a version"))
		}
		v, err2 := strconv.Atoi(os.Args[2])
		if err2 != nil {
			fail(err2)
		}
		err = m.Force(v)
	default:
		fail(fmt.Errorf("unknown command %q", os.Args[1]))
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fail(err)
	}
	fmt.Println("ok")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
