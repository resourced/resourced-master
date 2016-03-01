package migrator

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/libtime"
)

func New(generalConfig config.GeneralConfig) *Migrator {
	m := &Migrator{}
	m.AppGeneralConfig = generalConfig

	return m
}

type Migrator struct {
	AppGeneralConfig config.GeneralConfig
}

func (m *Migrator) CoreMigrateUp() ([]error, bool) {
	return migrate.UpSync(m.AppGeneralConfig.DSN, "./migrations/core")
}

func (m *Migrator) TSWatchersMigrateUp() ([]error, bool) {
	for _, dsn := range m.AppGeneralConfig.Watchers.DSNs {
		errs, ok := migrate.UpSync(dsn, "./migrations/ts-watchers")
		if errs != nil && len(errs) > 0 {
			return errs, ok
		}
	}

	return nil, true
}

func (m *Migrator) TSMetricsMigrateUp() ([]error, bool) {
	for _, dsn := range m.AppGeneralConfig.Metrics.DSNs {
		errs, ok := migrate.UpSync(dsn, "./migrations/ts-metrics")
		if errs != nil && len(errs) > 0 {
			return errs, ok
		}
	}

	return nil, true
}

func (m *Migrator) TSEventsMigrateUp() ([]error, bool) {
	for _, dsn := range m.AppGeneralConfig.Events.DSNs {
		errs, ok := migrate.UpSync(dsn, "./migrations/ts-events")
		if errs != nil && len(errs) > 0 {
			return errs, ok
		}
	}

	return nil, true
}

func (m *Migrator) CreateNextYearMigrationFiles() error {
	nextYear := time.Now().UTC().Year() + 1

	files, err := ioutil.ReadDir("./migrations/core")
	if err != nil {
		return err
	}

	// Don't do anything if next year migration files are already created
	for _, file := range files {
		if strings.Contains(file.Name(), fmt.Sprintf("%v", nextYear)) {
			return nil
		}
	}

	lastFile := files[len(files)-1]
	lastIndexString := strings.Split(lastFile.Name(), "_")[0]
	lastIndex, err := strconv.ParseInt(lastIndexString, 10, 64)
	if err != nil {
		return err
	}

	createMigrationCommands := []string{
		`cp -r ./migrations/core/0015_add-ts-watchers-2016.up.sql ./migrations/core/%04d_add-ts-watchers-%v.up.sql && sed -i'' -e "s/2016/%v/g" ./migrations/core/%04d_add-ts-watchers-%v.up.sql`,
		`cp -r ./migrations/core/0015_add-ts-watchers-2016.down.sql ./migrations/core/%04d_add-ts-watchers-%v.down.sql && sed -i'' -e "s/2016/%v/g" ./migrations/core/%04d_add-ts-watchers-%v.down.sql`,
		`cp -r ./migrations/ts-watchers/0015_add-ts-watchers-2016.up.sql ./migrations/ts-watchers/%04d_add-ts-watchers-%v.up.sql && sed -i'' -e "s/2016/%v/g" ./migrations/ts-watchers/%04d_add-ts-watchers-%v.up.sql`,
		`cp -r ./migrations/ts-watchers/0015_add-ts-watchers-2016.down.sql ./migrations/ts-watchers/%04d_add-ts-watchers-%v.down.sql && sed -i'' -e "s/2016/%v/g" ./migrations/ts-watchers/%04d_add-ts-watchers-%v.down.sql`,

		`cp -r ./migrations/core/0019_add-ts-metrics-2016.up.sql ./migrations/core/%04d_add-ts-metrics-%v.up.sql && sed -i'' -e "s/2016/%v/g" ./migrations/core/%04d_add-ts-metrics-%v.up.sql`,
		`cp -r ./migrations/core/0019_add-ts-metrics-2016.down.sql ./migrations/core/%04d_add-ts-metrics-%v.down.sql && sed -i'' -e "s/2016/%v/g" ./migrations/core/%04d_add-ts-metrics-%v.down.sql`,
		`cp -r ./migrations/ts-metrics/0019_add-ts-metrics-2016.up.sql ./migrations/ts-metrics/%04d_add-ts-metrics-%v.up.sql && sed -i'' -e "s/2016/%v/g" ./migrations/ts-metrics/%04d_add-ts-metrics-%v.up.sql`,
		`cp -r ./migrations/ts-metrics/0019_add-ts-metrics-2016.down.sql ./migrations/ts-metrics/%04d_add-ts-metrics-%v.down.sql && sed -i'' -e "s/2016/%v/g" ./migrations/ts-metrics/%04d_add-ts-metrics-%v.down.sql`,

		`cp -r ./migrations/core/0021_add-ts-events-2016.up.sql ./migrations/core/%04d_add-ts-events-%v.up.sql && sed -i'' -e "s/2016/%v/g" ./migrations/core/%04d_add-ts-events-%v.up.sql`,
		`cp -r ./migrations/core/0021_add-ts-events-2016.down.sql ./migrations/core/%04d_add-ts-events-%v.down.sql && sed -i'' -e "s/2016/%v/g" ./migrations/core/%04d_add-ts-events-%v.down.sql`,
		`cp -r ./migrations/ts-events/0021_add-ts-events-2016.up.sql ./migrations/ts-events/%04d_add-ts-events-%v.up.sql && sed -i'' -e "s/2016/%v/g" ./migrations/ts-events/%04d_add-ts-events-%v.up.sql`,
		`cp -r ./migrations/ts-events/0021_add-ts-events-2016.down.sql ./migrations/ts-events/%04d_add-ts-events-%v.down.sql && sed -i'' -e "s/2016/%v/g" ./migrations/ts-events/%04d_add-ts-events-%v.down.sql`,
	}

	index := lastIndex

	for i, command := range createMigrationCommands {
		if i%4 == 0 {
			index = index + int64(1)
		}

		cmd := fmt.Sprintf(command, index, nextYear, nextYear, index, nextYear)
		_, err = exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			return err
		}

		cmdChunks := strings.Split(cmd, " ")
		targetFile := cmdChunks[len(cmdChunks)-1]

		// Make sure we account for leap year when generating February end date.
		if !libtime.IsLeapYear(nextYear) {
			cmd = fmt.Sprintf(`sed -i'' -e "s/-02-29/-02-28/g" ` + targetFile)
			_, err = exec.Command("bash", "-c", cmd).CombinedOutput()
			if err != nil {
				return err
			}
		}
	}

	_, err = exec.Command("bash", "-c", `rm -rf ./migrations/**/*.sql-e`).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
