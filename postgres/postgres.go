package postgres

import "fmt"

type PostgresInfo struct {
	DatabaseName        string
	Server              string
	Port                string
	Username            string
	TempLocalBackupName string
}

func (p PostgresInfo) ToCommandOptions() []string {
	return []string{
		"-Fc",
		fmt.Sprintf("-d%s", p.DatabaseName),
		fmt.Sprintf("-h%s", p.Server),
		fmt.Sprintf("-p%s", p.Port),
		fmt.Sprintf("-U%s", p.Username),
		fmt.Sprintf("-f%s", p.TempLocalBackupName),
	}
}
