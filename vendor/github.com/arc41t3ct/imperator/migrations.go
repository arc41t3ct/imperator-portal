package imperator

import (
	_ "github.com/go-sql-driver/mysql"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (i *Imperator) MigrateUp(dsn string) error {
	m, err := migrate.New("file://"+i.RootPath+"/migrations", dsn)
	if err != nil {
		i.ErrorLog.Println("error getting migration files:", err)
		return err
	}
	// close after we are done
	defer m.Close()
	if err = m.Up(); err != nil {
		i.ErrorLog.Println("error running migrate up:", err)
	}
	return nil
}

func (i *Imperator) MigrateDownAll(dsn string) error {
	m, err := migrate.New("file://"+i.RootPath+"/migrations", dsn)
	if err != nil {
		i.ErrorLog.Println("error getting migration files:", err)
		return err
	}
	// close after we are done
	defer m.Close()
	if err := m.Down(); err != nil {
		i.ErrorLog.Println("error running migration down all:", err)
	}
	return nil
}

func (i *Imperator) Steps(n int, dsn string) error {
	m, err := migrate.New("file://"+i.RootPath+"/migrations", dsn)
	if err != nil {
		i.ErrorLog.Println("error getting migration files:", err)
		return err
	}
	// close after we are done
	defer m.Close()
	if err := m.Steps(n); err != nil {
		i.ErrorLog.Println("error running migration steps:", err)
		return err
	}
	return nil
}

func (i *Imperator) MigrateForce(dsn string) error {
	m, err := migrate.New("file://"+i.RootPath+"/migrations", dsn)
	if err != nil {
		i.ErrorLog.Println("error getting migration files:", err)
		return err
	}
	// close after we are done
	defer m.Close()
	if err := m.Force(-1); err != nil {
		i.ErrorLog.Println("error forcing migration step one down:", err)
		return err
	}
	return nil
}
