package postgres

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
)

type Db struct {
	Db *sqlx.DB
	Repo
}

type Repo interface {
	DeviceRepo
	UserRepo
	Close()
}

func NewDb(log *zap.SugaredLogger, conf *config.Config) *Db {
	pg_con_string := fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		conf.DB.Port, conf.DB.Hostname, conf.DB.Username, conf.DB.Password, conf.DB.DatabaseName)

	db, err := sqlx.Open("postgres", pg_con_string)
	if err != nil {
		log.Error(err)
		panic("Error when setting Database")
	}

	//driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	//if err != nil {
	//	log.Error(err)
	//	panic("Error when migrate DB")
	//}
	//
	//migrationsPath := conf.DB.MigrationsSourceURL
	//m, err := migrate.NewWithDatabaseInstance(migrationsPath, conf.DB.DatabaseName, driver)
	//if err != nil {
	//	log.Error(err)
	//	panic("Error when migrate DB")
	//}
	//
	//err = m.Up()
	//if err != nil && !errors.Is(err, migrate.ErrNoChange) {
	//	log.Error(err)
	//	panic("Error when migrate DB")
	//}

	return &Db{Db: db}
}

func (d *Db) Close() {
	d.Db.Close()
}
