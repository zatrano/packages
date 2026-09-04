package database

import (
	"database/sql"
	"strings"

	appconfig "github.com/zatrano/framework/config"
	pkgconfig "github.com/zatrano/framework/config"
	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/packages/bootutil"
	"github.com/zatrano/packages/database/query"
	"github.com/zatrano/packages/events"
	"github.com/zatrano/packages/orm"
	"github.com/zatrano/packages/validation"
)

func boot(app contracts.App) error {
	pkgconfig.LoadIfAbsent(app.Config(), "database", appconfig.Database())
	defaultConn := strings.TrimSpace(app.Config().GetString("database.default"))
	connections := map[string]ConnectionConfig{}

	rawConnections, ok := app.Config().Get("database.connections").(map[string]any)
	if !ok {
		rawConnections = map[string]any{}
	}

	for name, raw := range rawConnections {
		cfgMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		cfg := ConnectionConfig{
			Driver:   bootutil.AsString(cfgMap["driver"]),
			Host:     bootutil.AsString(cfgMap["host"]),
			Port:     bootutil.AsString(cfgMap["port"]),
			Database: bootutil.AsString(cfgMap["database"]),
			Username: bootutil.AsString(cfgMap["username"]),
			Password: bootutil.AsString(cfgMap["password"]),
			Charset:  bootutil.AsString(cfgMap["charset"]),
			SSLMode:  bootutil.AsString(cfgMap["sslmode"]),
			Service:  bootutil.AsString(cfgMap["service"]),
			URI:      bootutil.AsString(cfgMap["uri"]),
		}
		if IsDocumentStore(cfg.Driver) || IsDocumentStore(name) {
			// Document stores are provided by github.com/zatrano/packages/mongo.
			continue
		}
		connections[name] = cfg
	}

	if len(connections) == 0 {
		return nil
	}

	sqlDefault := defaultConn
	if _, ok := connections[sqlDefault]; !ok {
		for name := range connections {
			sqlDefault = name
			break
		}
	}

	mgr := NewManager(Config{
		Default:     sqlDefault,
		Connections: connections,
	}, app.BasePath())
	app.Container().Instance("db", mgr)

	db, err := mgr.DB()
	if err != nil {
		return err
	}
	driver, err := mgr.DriverName()
	if err != nil {
		return err
	}
	orm.Configure(db, driver)
	orm.SetConnectionResolver(func(name string) (*sql.DB, string, error) {
		conn, err := mgr.Connection(name)
		if err != nil {
			return nil, "", err
		}
		d, err := mgr.DriverName(name)
		if err != nil {
			return nil, "", err
		}
		return conn, d, nil
	})
	if enc := app.Encrypter(); enc != nil {
		orm.SetCastEncrypter(enc)
	}
	if ev := events.From(app); ev != nil {
		orm.SetDispatcher(ev)
	}
	validation.SetDefaultPresenceChecker(func(table, column, value string) (bool, error) {
		table = strings.TrimSpace(table)
		column = strings.TrimSpace(column)
		if table == "" || column == "" {
			return false, nil
		}
		row, err := query.New(db, driver, table).Where(column, value).First()
		if err != nil {
			if err == sql.ErrNoRows {
				return false, nil
			}
			return false, err
		}
		return row != nil, nil
	})
	return nil
}
